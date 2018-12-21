#!/usr/bin/env bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
#

#
#  This script initiates batch runs on one or more remote marbles-perf servers
#  and keeps polling for the results until all batches are complete.
#
#  Args:
#    $1 - mandatory, concurrency size of each batch job (number of marbles to create per server)
#    $2 - optional, iterations, ie. number of transfers to perform per marbles, default is 50
#    $3 - optional, size of extra data to be added to marbles to affect size of fabric transactions, default is 20
#
#  Envs:
#    MARBLE_APP_SERVERS - optional, a space-separated list of base server URLs to participate in the run
#                         If this environment is not set, http://localhost:8080 will be used.
#
#    MARBLE_POLL_INTERVAL - optional, how often do we poll server for results in seconds, default is 60
#

concurrency=$1
iterations=${2:-50}
extraDataLength=${3:-20}

server_list=${MARBLE_APP_SERVERS:-"http://localhost:8080"}
poll_interval=${MARBLE_POLL_INTERVAL:-60}

if [ -z "$concurrency" ] ; then
    echo missing concurrency
    exit 1
fi

tmp_request_file=/tmp/marbles_request_$$.json

cat <<END_REQUEST > $tmp_request_file
    {"concurrency":$concurrency, "iterations":$iterations, "clearMarbles":true, "extraDataLength":$extraDataLength}
END_REQUEST

echo $(date) start new test ...

echo .
echo test parameters:
cat $tmp_request_file

echo .
echo server list: $server_list

servers=( $server_list )

# create users o1 to o5 for testing
owner_url=${servers[0]}/owner
for owner_idx in $(seq 1 10) ; do
    # check if owner exists
    if [ $(curl -I -X GET $owner_url/o${owner_idx} 2> /dev/null | grep -c '200 OK') -eq 0 ] ; then
        echo "user o${owner_idx} does not exist, will try to create ..."
        owner_data="{\"id\":\"o${owner_idx}\",\"username\":\"user${owner_idx}\",\"company\":\"company_${owner_idx}\"}"
        curl -X POST -d $owner_data $owner_url 2> /dev/null
        echo
    fi
done


# initiate batch runs
#
for server in $server_list ; do
    request_file=$tmp_request_file
    echo curl -q -X POST -d @$request_file ${server}/batch_run
    resp=$( curl -q -X POST -d @$request_file ${server}/batch_run 2> /dev/null )
    curl_exit_code=$?
    batch_id=""
    if [ $curl_exit_code -eq 0 ] ; then
        batch_id=$(echo $resp | awk '/batchId/{gsub("\"", "", $3) ; print $3}')
    fi
    if [ -z "$batch_id" ] ; then
        echo failed to start batch run at server $server
        echo $resp
        exit 1
    fi
    echo batch run started on $server, batch_id: $batch_id
    BATCH_ID_LIST="$BATCH_ID_LIST $batch_id"
done

#
# initialize batch results as empty
#
batch_ids=( $BATCH_ID_LIST)
declare -a batch_results

# total number of successful transfers across all batch runs
let transfer_count=0

# total number of seconds spent on successful transfers
let transfer_seconds=0

echo
echo batch runs initiated, will poll server for results every $poll_interval seconds...
echo

#
# poll server until all batch runs are done
#
while [ 0 ] ; do
    sleep $poll_interval
    echo
    echo $(date) poll server for results ...
    echo
    let complete_count=0
    for i in ${!batch_ids[*]} ; do
        if [ -z "${batch_results[${i}]}" ] ; then
            batch_id=${batch_ids[$i]}
            echo curl -q ${servers[0]}/batch_run/$batch_id
            resp=$(curl -q ${servers[0]}/batch_run/$batch_id 2> /dev/null)
            if [ $? -eq 0 ] && [ $( echo $resp | grep -c 'totalSuccesses' ) -gt 0 ] ; then
                batch_results[${i}]="$resp"
            fi
        fi
        if [ ! -z "${batch_results[${i}]}" ] ; then
            let complete_count=$complete_count+1
        fi
    done
    if [ $complete_count == ${#batch_ids[*]} ] ; then
        break
    fi
    echo
    echo $(date) progress: $complete_count out of ${#batch_ids[*]} servers completed, will check again in $poll_interval seconds
    echo
done


#
# complete - clean up temp files
#
rm -f $tmp_request_file


#
# display results
#
echo ALL batch runs complete:
for result in "${batch_results[@]}" ; do
    echo
    echo $result
    echo
    if [ $(echo ${result} | grep -c totalSuccesses) -gt 0 ] ; then
            let successes=$(echo ${result} | tr '}' '\n' | tr ',' '\n' | awk -v FS=: '/totalSuccesses/{gsub(" ","",$2); gsub("}","",$2); print $2}' )
            let successSecs=$(echo ${result} | tr '}' '\n' | tr ',' '\n' | awk -v FS=: '/totalSuccessSeconds/{gsub(" ","",$2); gsub("}","",$2); print $2}' )
            let transfer_seconds+=$successSecs
            let transfer_count+=$successes
    fi
done

echo
echo
echo Performance Summary:
echo
echo    Combined total number of transfers: $transfer_count
echo    Total time in seconds on transfers: $transfer_seconds
echo
if [ $transfer_count -gt 0 ] ; then
   echo    Combined average time in seconds per transfer: $( echo "scale=3; $transfer_seconds/$transfer_count" | bc ) 
   echo
   min_transfer_time=$( echo ${batch_results[@]} | tr '}' '\n' | tr ',' '\n' | awk -v FS=: '/minTransferSeconds/{gsub(" ","",$2); gsub("}","",$2); print $2}' | sort -n | head -1)
   max_transfer_time=$( echo ${batch_results[@]} | tr '}' '\n' | tr ',' '\n' | awk -v FS=: '/maxTransferSeconds/{gsub(" ","",$2); gsub("}","",$2); print $2}' | sort -n | tail -1)
   echo min transfer time in seconds: $min_transfer_time
   echo max transfer time in seconds: $max_transfer_time
   echo
   echo
fi




