#!/bin/bash
#
# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
#
# http:#www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.
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

