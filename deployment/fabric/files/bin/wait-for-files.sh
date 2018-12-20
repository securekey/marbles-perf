#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
#

#
# wait_for_file - waits for 1 file
#
function wait_for_file() {

    file="$1"

    echo "wait-for-file: $file"
    while [ ! -r "$file" ] ; do 
        echo "$(date '+%Y-%m-%d %H:%M:%S') - waiting for file...  $file"
        sleep 2 ; 
    done
}


#
# main - wait for 1 or more files
#

args=("$@")

echo "wait for files: $args"

for ((i=0;i<${#args[@]};i+=1))
do
    wait_for_file "${args[$i]}"
done

