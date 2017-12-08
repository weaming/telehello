#!/bin/bash

PWD=`pwd`
outfile=$PWD/crontab
cat > ${outfile} << EOF
0 8,18 * * * python $PWD/BlockChainsPrice.py
EOF
crontab ${outfile} && crontab -l
