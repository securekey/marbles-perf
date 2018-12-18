#!/usr/bin/env bash
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
#                         if more than 1 server is provided, all servers except for the last one
#                         in the list will run 15 more iterations than the specified value of iterations.
#                         This is to garantee the load of the overall system for the duration of entire run of
#                         the last server so that the average transfer time reported by the last server, is
#                         the average under a consistent load concurrent threads = concurrency x number of servers.
#                         If this environment is not set, http://localhost:8080 will be used.
#
#

concurrency=$1
iterations=${2:-50}
extraDataLength=${3:-20}

server_list=${MARBLE_APP_SERVERS:-"http://localhost:8080"}

if [ -z "$concurrency" ] ; then
    echo missing concurrency
    exit 1
fi

tmp_request_file=/tmp/marbles_request_$$.json
tmp_load_file=/tmp/marbles_load_$$.json

let load_iterations=$iterations+15

cat <<EOF > $tmp_load_file
    {"concurrency":$concurrency, "iterations":$load_iterations, "clearMarbles":false, "extraDataLength":$extraDataLength}
EOF

cat <<END_REQUEST > $tmp_request_file
    {"concurrency":$concurrency, "iterations":$iterations, "clearMarbles":false, "extraDataLength":$extraDataLength}
END_REQUEST

echo $(date) start new test ...

echo .
echo load parameters:
cat $tmp_load_file

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


# clear marbles first
echo
echo clearing all existing marbles in system
curl -X POST ${servers[0]}/clear_marbles
echo


# initiate batch runs
#
let request_count=0
for server in $server_list ; do
    let request_count=request_count+1
    if [ $request_count == ${#servers[*]} ] ; then
        # last server
        request_file=$tmp_request_file
    else
        request_file=$tmp_load_file
    fi

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
for ix in ${!batch_ids[*]} ; do
    batch_results[${ix}]=""
done

#
# poll server until all batch runs are done
#
while [ 0 ] ; do
    sleep 60
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
        echo ALL batch runs complete:
        for result in "${batch_results[@]}" ; do
            echo
            echo $result
            echo
        done
        break
    fi
    echo
    echo $(date) progress: $complete_count out of ${#batch_ids[*]} servers completed, will check again in 60 seconds
    echo
done








