package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"math/rand"
	"os"
//	"path/filepath"
	"time"
	"bufio"
	"github.com/ipfs/go-cid"
	"strconv"

	saproof "github.com/filecoin-project/specs-actors/actors/runtime/proof"

	"github.com/docker/go-units"
	logging "github.com/ipfs/go-log/v2"
//	"github.com/minio/blake2b-simd"
	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"

	"github.com/filecoin-project/go-address"
	paramfetch "github.com/filecoin-project/go-paramfetch"
	"github.com/filecoin-project/go-state-types/abi"
	lcli "github.com/filecoin-project/lotus/cli"
	"github.com/filecoin-project/lotus/extern/sector-storage/ffiwrapper"
	"github.com/filecoin-project/lotus/extern/sector-storage/ffiwrapper/basicfs"
//	"github.com/filecoin-project/lotus/extern/sector-storage/stores"
	"github.com/filecoin-project/specs-actors/actors/builtin/miner"
//	"github.com/filecoin-project/specs-storage/storage"

//	lapi "github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/build"
	"github.com/filecoin-project/lotus/chain/types"
)

var log = logging.Logger("lotus-bench")

type Commit2In struct {
        SectorNum  int64
        Phase1Out  []byte
        SectorSize uint64
}


func main() {
	logging.SetLogLevel("*", "INFO")

	log.Info("Starting lotus-bench")

	miner.SupportedProofTypes[abi.RegisteredSealProof_StackedDrg2KiBV1] = struct{}{}

	app := &cli.App{
		Name:    "sector-sanity-check",
		Usage:   "check window post",
		Version: build.UserVersion(),
		Commands: []*cli.Command{
			proveCmd,
			sealBenchCmd,
			importBenchCmd,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Warnf("%+v", err)
		return
	}
}

var sealBenchCmd = &cli.Command{
	Name: "checking",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "storage-dir",
			Value: "~/.lotus-bench",
			Usage: "Path to the storage directory that will store sectors long term",
		},
		&cli.StringFlag{
			Name:  "sector-size",
			Value: "512MiB",
			Usage: "size of the sectors in bytes, i.e. 32GiB",
		},
                &cli.StringFlag{
                        Name:  "sectors-file",
                        Value: "sectors.txt",
                        Usage: "absolute path file. contains line number, line cidcommr, line number...",
                },
		&cli.BoolFlag{
			Name:  "no-gpu",
			Usage: "disable gpu usage for the checking",
		},
		&cli.StringFlag{
			Name:  "miner-addr",
			Usage: "pass miner address (only necessary if using existing sectorbuilder)",
			Value: "t010010",
		},
                &cli.IntFlag{
                        Name:  "number",
                        Value: 1,
                },

                &cli.StringFlag{
                        Name:  "cidcommr",
			Usage: "CIDcommR,  eg/default.  bagboea4b5abcbkyyzhl37s5kyjjegeysedpczhija7cczazapavjejbppck57b2z",
			Value: "bagboea4b5abcbkyyzhl37s5kyjjegeysedpczhija7cczazapavjejbppck57b2z",
                },
	},
	Action: func(c *cli.Context) error {
		if c.Bool("no-gpu") {
			err := os.Setenv("BELLMAN_NO_GPU", "1")
			if err != nil {
				return xerrors.Errorf("setting no-gpu flag: %w", err)
			}
		}

		var sbdir string

		sdir, err := homedir.Expand(c.String("storage-dir"))
		if err != nil {
			return err
		}

		err = os.MkdirAll(sdir, 0775) //nolint:gosec
		if err != nil {
			return xerrors.Errorf("creating sectorbuilder dir: %w", err)
		}

		defer func() {
		}()

		sbdir = sdir

		// miner address
		maddr, err := address.NewFromString(c.String("miner-addr"))
		if err != nil {
			return err
		}
		log.Infof("miner maddr: ", maddr)
		amid, err := address.IDFromAddress(maddr)
		if err != nil {
			return err
		}
		log.Infof("miner amid: ", amid)
		mid := abi.ActorID(amid)
		log.Infof("miner mid: ", mid)

		// sector size
		sectorSizeInt, err := units.RAMInBytes(c.String("sector-size"))
		if err != nil {
			return err
		}
		sectorSize := abi.SectorSize(sectorSizeInt)

		spt, err := ffiwrapper.SealProofTypeFromSectorSize(sectorSize)
		if err != nil {
			return err
		}

		cfg := &ffiwrapper.Config{
			SealProofType: spt,
		}

		if err := paramfetch.GetParams(lcli.ReqContext(c), build.ParametersJSON(), uint64(sectorSize)); err != nil {
			return xerrors.Errorf("getting params: %w", err)
		}

		sbfs := &basicfs.Provider{
			Root: sbdir,
		}

		sb, err := ffiwrapper.New(sbfs, cfg)
		if err != nil {
			return err
		}


		sealedSectors := getSectorsInfo(c.String("sectors-file"), sb.SealProofType())

		var challenge [32]byte
		rand.Read(challenge[:])

		log.Info("computing window post snark (cold)")
		wproof1, ps, err := sb.GenerateWindowPoSt(context.TODO(), mid, sealedSectors, challenge[:])
		if err != nil {
			return err
		}

		wpvi1 := saproof.WindowPoStVerifyInfo{
			Randomness:        challenge[:],
			Proofs:            wproof1,
			ChallengedSectors: sealedSectors,
			Prover:            mid,
		}

		log.Info("generate window PoSt skipped sectors", "sectors", ps, "error", err)

		ok, err := ffiwrapper.ProofVerifier.VerifyWindowPoSt(context.TODO(), wpvi1)
		if err != nil {
			return err
		}
		if !ok {
			log.Error("post verification failed")
		}

		return nil
	},
}

func getSectorsInfo(filePath string, proofType abi.RegisteredSealProof) []saproof.SectorInfo {

	sealedSectors := make([]saproof.SectorInfo, 0)

	file, err := os.Open(filePath)
	defer file.Close()
	if err != nil {
		return sealedSectors
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		sectorIndex := scanner.Text()

		index, error := strconv.Atoi(sectorIndex)
		if error != nil {
			fmt.Println("error")
			break
		}

		scanner.Scan()
		cidStr := scanner.Text()
		ccid, err := cid.Decode(cidStr)
                if(err != nil) {
                        log.Infof("cid error, ignore sectors after this: %d, %s", uint64(index), err)
			return sealedSectors 
                }

		var sector saproof.SectorInfo
		sector.SealProof = proofType
		sector.SectorNumber = abi.SectorNumber(uint64(index))
		sector.SealedCID = ccid

		sealedSectors = append(sealedSectors, sector)

		log.Infof("id: ", sector.SectorNumber)
		log.Infof("cid: ", sector.SealedCID)

	}

	fmt.Println("sector length", len(sealedSectors))
	return sealedSectors
}

type ParCfg struct {
	PreCommit1 int
	PreCommit2 int
	Commit     int
}


var proveCmd = &cli.Command{
	Name:  "prove",
	Usage: "Benchmark a proof computation",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "no-gpu",
			Usage: "disable gpu usage for the benchmark run",
		},
		&cli.StringFlag{
			Name:  "miner-addr",
			Usage: "pass miner address (only necessary if using existing sectorbuilder)",
			Value: "t01000",
		},
	},
	Action: func(c *cli.Context) error {
		if c.Bool("no-gpu") {
			err := os.Setenv("BELLMAN_NO_GPU", "1")
			if err != nil {
				return xerrors.Errorf("setting no-gpu flag: %w", err)
			}
		}

		if !c.Args().Present() {
			return xerrors.Errorf("Usage: lotus-bench prove [input.json]")
		}

		inb, err := ioutil.ReadFile(c.Args().First())
		if err != nil {
			return xerrors.Errorf("reading input file: %w", err)
		}

		var c2in Commit2In
		if err := json.Unmarshal(inb, &c2in); err != nil {
			return xerrors.Errorf("unmarshalling input file: %w", err)
		}

		if err := paramfetch.GetParams(lcli.ReqContext(c), build.ParametersJSON(), c2in.SectorSize); err != nil {
			return xerrors.Errorf("getting params: %w", err)
		}

		maddr, err := address.NewFromString(c.String("miner-addr"))
		if err != nil {
			return err
		}
		mid, err := address.IDFromAddress(maddr)
		if err != nil {
			return err
		}

		spt, err := ffiwrapper.SealProofTypeFromSectorSize(abi.SectorSize(c2in.SectorSize))
		if err != nil {
			return err
		}

		cfg := &ffiwrapper.Config{
			SealProofType: spt,
		}

		sb, err := ffiwrapper.New(nil, cfg)
		if err != nil {
			return err
		}

		start := time.Now()

		proof, err := sb.SealCommit2(context.TODO(), abi.SectorID{Miner: abi.ActorID(mid), Number: abi.SectorNumber(c2in.SectorNum)}, c2in.Phase1Out)
		if err != nil {
			return err
		}

		sealCommit2 := time.Now()

		fmt.Printf("proof: %x\n", proof)

		fmt.Printf("----\nresults (v27) (%d)\n", c2in.SectorSize)
		dur := sealCommit2.Sub(start)

		fmt.Printf("seal: commit phase 2: %s (%s)\n", dur, bps(abi.SectorSize(c2in.SectorSize), dur))
		return nil
	},
}

func bps(data abi.SectorSize, d time.Duration) string {
	bdata := new(big.Int).SetUint64(uint64(data))
	bdata = bdata.Mul(bdata, big.NewInt(time.Second.Nanoseconds()))
	bps := bdata.Div(bdata, big.NewInt(d.Nanoseconds()))
	return types.SizeStr(types.BigInt{Int: bps}) + "/s"
}
