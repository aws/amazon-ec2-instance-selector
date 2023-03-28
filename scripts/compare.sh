#!/bin/bash

## This script helps compare 2 EC2 instance types
## dependencies: 
##     1. vimdiff: https://command-not-found.com/vimdiff
##     2. EC2 instance selector CLI: https://github.com/aws/amazon-ec2-instance-selector

ec2-instance-selector -r us-west-2 -v --allow-list=$1 > /tmp/$1 2>/dev/null
ec2-instance-selector -r us-west-2 -v --allow-list=$2 > /tmp/$2 2>/dev/null

vimdiff /tmp/$1 /tmp/$2
