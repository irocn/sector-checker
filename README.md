## sector-sanity-checker-v0.3.0

This tools can help you check sector to avoid the window PoST fail.

## Download

https://github.com/irocn/sector-checker/releases/tag/sector-sanity-checker-v0.3.0

## Usage
### step 1, export the environment variable
 - export FIL_PROOFS_PARENT_CACHE=<YOUR_PARENT_CACHE>
 - export FIL_PROOFS_PARAMETER_CACHE=<YOUR_FIL_PROOFS_PARAMETER_CACHE>
 - export FIL_PROOFS_USE_GPU_COLUMN_BUILDER=1 
 - export RUST_LOG=info FIL_PROOFS_USE_GPU_TREE_BUILDER=1 
 - export FIL_PROOFS_MAXIMIZE_CACHING=1
 - export MINER_API_INFO=<YOUR_MINER_API_INFO>
### step 2, run the tool 
 - $>sector-sanity-checker checking  --sector-size=32G --miner-addr=<your_miner_id> --storage-dir=<sector_dir> 
 - $>sector-sanity-checker checking  --sector-size=32G --sectors-file-only-number=<sectors-to-scan> --miner-addr=<your_miner_id> --storage-dir=<sector_dir>
 
### For Example:

 - $>sector-sanity-checker checking  --sector-size=32G --miner-addr=t### --storage-dir=/opt/data/storage
 
 Then all the sectors under /opt/data/storage/sealed/s-xxxxx-xxx will be scaned.
 
 - $>sector-sanity-checker checking  --sector-size=32G --sectors-file-only-number=sectors-to-scan.txt --miner-addr=t### --storage-dir=/opt/data/storage
 
 Then all the sectors specified by sectors-to-scan.txt  under folder /opt/data/storage will be scaned. 
   The file sectors-to-scan.txt contains the sector numbers to be scaned, each number has one line.
   The folder /opt/data/storage contains folder sealed and cache
 
  
![image](https://github.com/irocn/sector-sanity-checker/blob/master/1599813675963.jpg)

![image](https://github.com/irocn/sector-sanity-checker/blob/master/Screen%20Shot%202020-09-12%20at%2002.01.47.png)
## donate
If the tool help you, please donate 1 FIL to us.
 - Wallet Address: t3qtvkskn35hjj4sg2r3ce2j7x3arqcv7nexhhzthktfhpslc4agpkdq434kf5xh64nkzl7mix5cexwayhtgja  
