while IFS= read -r line
do
./lotus-miner sectors status --log $line | sed -n '1p;4p' | cut -d ':' -f2 | tr -s "\t"
done < "./sector-id.txt"
