#!/usr/bin/env bash
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
# detect changes in Gopkg.lock and run godep to populate vendor only if necessary
# output checksum to .cache/vendor.log for future detection
#

logfile=.cache/vendor.log

old_checksum=old
new_checksum=new

if [ -r $logfile ] ; then
    old_checksum=$(cat $logfile)
fi

if [ $(which cksum | wc -l) -gt 0 ] ; then
    new_checksum=$(cksum Gopkg.lock)
else
    echo cksum command not found, cannot cache vendoring
fi

if [ "$old_checksum" == "$new_checksum" ]; then
    echo vendor is up-to-date, no vendoring needed.  To force vendoring regardless, remove $logfile file
else
    dep ensure --vendor-only -v
    dep_exit_code=$?
    if [ $dep_exit_code -eq 0 ] && [ $(which cksum | wc -l) -gt 0 ] ; then
        mkdir -p .cache
        echo $new_checksum > $logfile
    fi
    exit $dep_exit_code
fi
