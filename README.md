# sector-sanity-checker

This tools can help you check sector to avoid the window PoST fail.

# Download

https://github.com/irocn/sector-sanity-checker/releases/tag/v0.0.1

# Usage
FIL_PROOFS_PARENT_CACHE=<YOUR_PARENT_CACHE>
FIL_PROOFS_PARAMETER_CACHE=<YOUR_FIL_PROOFS_PARAMETER_CACHE>
FIL_PROOFS_USE_GPU_COLUMN_BUILDER=1 
RUST_LOG=info FIL_PROOFS_USE_GPU_TREE_BUILDER=1 
FIL_PROOFS_MAXIMIZE_CACHING=1
$>sector-sanity-checker sealing --cidcommr=<sector_cidcommr>  --number=<sector_id> --sector-size=32GiB --miner-addr=<your_miner_id> --storage-dir=<sector_dir> 

For Example:

$>sector-sanity-checker sealing --cidcommr=bagboea4b5abcbkyyzhl37s5kyjjegeysedpczhija7cczazapavjejbppck57b2z --number=1000 --miner-addr=t### --sector-size=32GiB --storage-dir=/opt/data/storage

You may use lotus-miner sectors status --log <sector-id> to find the --cidcommr. 

# donate
If the tool help you, please donate 1 FIL to us.
Wallet Address: t3qtvkskn35hjj4sg2r3ce2j7x3arqcv7nexhhzthktfhpslc4agpkdq434kf5xh64nkzl7mix5cexwayhtgja  
