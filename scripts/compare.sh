#!/bin/bash

# Loop through all the arguments passed to the script
for arg in "$@"
do
#    ec2-instance-selector -r us-west-2 -v --allow-list=$1 > /tmp/$1 2>/dev/null
    if [ ! -f "/tmp/$arg" ]; then
        # If the file doesn't exist, download
        ec2-instance-selector -r us-west-2 -v --allow-list=$arg --price-per-hour-min 0 > /tmp/$arg 2>/dev/null
    fi
done

pushd .

cd /tmp
vimdiff "$@"

popd
