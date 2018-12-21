# Purpose
The goal of this project is to provide a way to test the performance of hyperledger fabric using a simple chaincode.  The chaincode used in this project is a slight modification of the hyperledger example marbles chaincode.


# Building Docker Image

Execute the following make target to build a docker image for marbles application server:
marbles-perf.

```
make docker
```



# Setting Up Runtime Environment Locally
Follow these steps to set up a complete environment for testing locally.

## Start Fabric Network
Use commands below to spin up a Hyperledger Fabric network with 3 organizations with a total of 9 peers.  Marbles chancode is deployed with endorsement policy that requires any 2 of the 3 organizations.

```
# set up your local GOPATH
# build scripts expect marbles-perf project located at ${GOPATH}/src/github.com/marbles-perf

# change to suit to your local environment
export GOPATH=${HOME}/go

# stop all running containers and start up fabric network
docker stop $(docker ps -q) && \
make fabric-up
```


## Start marbles-perf web service
Use the make target below to start up Marbles web service:

```
make services-up
```


## Sanity Test of Runtime Environment
Optionally, one can run these commands as a quick sanity test on the runtime environment.

```
curl -X POST -d '{"id": "ojdoe1", "username": "jdoe1", "company": "jdoe1_corp"}' http://localhost:8080/owner && \
curl -X POST -d '{"id": "ojdoe2", "username": "jdoe2", "company": "jdoe2_corp"}' http://localhost:8080/owner && \
curl -X POST http://localhost:8080/clear_marbles && \
curl -X POST -d '{"id": "mUnittest2", "color": "blue", "size": 13, "owner": {"id":"ojdoe1", "company": "jdoe1_corp"}, "company": "jdoe1_corp"}' http://localhost:8080/marble && \
curl -X POST -d '{"id": "mUnittest3", "color": "yellow", "size": 20, "owner": {"id":"ojdoe1", "company": "jdoe1_corp"}, "company": "jdoe1_corp", "additionalData": "hello"}' http://localhost:8080/marble && \
curl -X POST -d '{"marbleId": "mUnittest3", "toOwnerId": "ojdoe2", "authCompany": "jdoe1_corp"}' http://localhost:8080/transfer && \
curl http://localhost:8080/marble/mUnittest3

```


# Running Marbles Performance Tests

The marbles-perf web service provide these endpoints for performance testing purposes:

## /batch_run
This endpoint initiates a performance run process in background which creates a set of threads to execute marble ownership transfers in parallel.  Exactly one marble is created at the beginning of each thread execution.  The newly created marble will be transferred from one user to another until the number of transfers reach the specified iterations.  The server assigns a batch_id for this process and returns this id within the HTTP response to the caller.  The caller can poll the server periodically for the results of the process using this id.

```
Endpoint: /batch_run
Method: POST
Request Payload:
{
   "concurrency": 100,
   "iterations": 50,
   "delaySeconds": 0,
   "clearMarbles": false
   "extraDataLength": 20
}

Response Payload:
{
	"batchId": "some unique string to fetch results later"
}
```

The meanings of the request JSON attributes are:

|Attribute|Meaning|
|-----------------|-------|
|concurrency|The number of concurrent workers|
|iterations|The number of transfers to be completed in each worker|
|delaySeconds|The number of seconds to wait between iterations|
|clearMarbles|Boolean indicating whether marbles created during this batch run should be deleted at the completion of the process|
|extraDataLength|Size of additional data to be added to each marble. This is provided to observe the effect of larger transactions on the ledger. Random data will be generated and added to each marble at create time. Subsequent transfers store the marble state so will also use the increased size.|


## /batch_run/{id}
This endpoint fetches results for a performance run.

### Request

```
Endpoint: /batch_run/{id}
Method: GET

```
The id is a batchId obtained from a previous call to /batch_run endpoint.


### Response
A response before the completion of the performance run looks like below:

```
HTTP Status: 404
Response Body:
{
	"error": "Batch run status not yet available (not complete)"
}
```

Response for a completed process would be like below:

```
{
  "request": {
    "concurrency": 5,
    "iterations": 65,
    "delaySeconds": 0,
    "clearMarbles": false,
    "extraDataLength": 20
  },
  "status": "success",
  "totalSuccesses": 325,
  "totalFailures": 0,
  "totalSuccessSeconds": 356,
  "averageTransferSeconds": 1.098,
  "minTransferSeconds": 1.03,
  "maxTransferSeconds": 1.57
}
```



# Running Performance On Remote Servers
A Bash script is provided for your convenience to start multiple performance loads on multiple servers and poll their results.
The location of the script is *scripts/start_load.sh*.

The header comment section of the script provides detailed description on its usage.  Below is a quick example:

```
$ export MARBLE_APP_SERVERS="http://marbles1.example.com http://marbles2.example.com http://marbles3.example.com"

$ export MARBLE_POLL_INTERVAL=60

# run 500 threads, 50 iterations, 30 bytes extra data
$ ./start_load.sh 500 50 30

# the script will start loads on all 3 servers and poll results every 60 seconds until completion.

```

