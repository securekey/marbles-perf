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

BROADCAST_MSG() {
  MESSAGE=$1
  echo
  echo "======================================================================="
  echo "    ${MESSAGE}"
  echo "======================================================================="
  echo
}

function DOCKER_NETWORK_SETUP() {
  DOCKER_HOST_ARG=$1
  echo "DOCKER_NETWORK_SETUP DOCKER_HOST_ARG is $DOCKER_HOST_ARG"

  bash -c "$DOCKER_HOST_ARG docker network create --driver=bridge --subnet=10.177.177.0/24 --gateway=10.177.177.1 --label=\"vme local dev compose network\" vme.sk.dev "

  # TODO: make this function nicer, catch errors
  # docker network create \
  #   --driver=bridge \
  #   --subnet=10.177.177.0/24 \
  #   --gateway=10.177.177.1 \
  #   --label="vme local dev compose network" \
  #   vme.sk.dev

}

#
# For use with source chaincode installation
#
function PREPARE_CHAINCODES() {

  rm -fr "$CC_BUILD_DIR"
  mkdir -p "$CC_BUILD_DIR"

  cc_src_dir=${COMPOSE_DIR}/chaincodes
  cp -rp $cc_src_dir $CC_BUILD_DIR

  echo "COPYING GO VENDORING for CC"
  for cc in $CC_BUILD_DIR/chaincodes/*_cc ; do
    if [ -d ${cc} ] ; then
      cp -rp $CC_BUILD_DIR/vendor ${cc}
    fi
  done

  echo "FIXING PERMISSIONS"
  for cc in $CC_BUILD_DIR/chaincodes/*_cc ; do
    find $cc -type f -exec chmod 664 {} \;
  done

  echo "DELETING BINARIES"
  for cc in marbles_cc consortium_cc custodian_cc dlbp_cc steward_cc ; do
    rm -rf $CC_BUILD_DIR/chaincodes/${cc}/${cc}
    rm -rf $CC_BUILD_DIR/chaincodes/${cc}/debug.test
  done
}

function NETWORK_DOWN() {  
    REMOVE_CONTAINERS $1
    REMOVE_COPIED_FILES
}

function REMOVE_CONTAINERS() {
  DOCKER_HOST_ARG=$1
  echo "REMOVE_CONTAINERS DOCKER_HOST_ARG is $DOCKER_HOST_ARG"

  bash -c "$DOCKER_HOST_ARG docker-compose ${COMPOSE_OPTS} down -v"
  bash -c "$DOCKER_HOST_ARG docker rm \$($DOCKER_HOST_ARG docker ps -a --filter name=\"dev-peer[0-9a-zA-Z_]*.vme.sk.dev-*\" -q)"
  bash -c "$DOCKER_HOST_ARG docker rmi \$($DOCKER_HOST_ARG docker images \"dev-peer[0-9a-zA-Z_]*.vme.sk.dev-*\" -q)"
}

function REMOVE_COPIED_FILES() {
  if [ "$MARBLESCC_KEEP_TEMP_FILES" != "true" ] && [ "$INSTALL_BINARY_CC" != "true" ] ; then
    rm -fr "$CC_BUILD_DIR"
  fi
}

function VALIDATE_ARGS () {
	if [ -z "${UP_DOWN}" ]; then
		echo "Option up / down / restart not mentioned"
		PRINT_HELP
		exit 1
	fi
}

function PRINT_HELP () {
	echo "Usage: ./network_setup <up|down> <pull|>"
}

function CALL_PREPARE_CHAINCODES() {
    if [ -z "$CC_BUILD_DIR" ] ; then
      export CC_BUILD_DIR=/tmp/marbles-cc-build/$USER
    fi
    export CORE_CHAINCODE_BUILDER=${FABRIC_CCENV_FIXTURE_IMAGE}:${FABRIC_CCENV_FIXTURE_TAG}
    export CORE_CHAINCODE_GOLANG_RUNTIME=${FABRIC_BASEOS_FIXTURE_IMAGE}:${FABRIC_BASEOS_FIXTURE_TAG}

    BROADCAST_MSG "PREPARING CHAINCODES"
    PREPARE_CHAINCODES
}

function NETWORK_UP() {
  DOCKER_HOST_ARG=$1
  BROADCAST_MSG "Docker Remote hostname is $DOCKER_HOST_ARG"

  BROADCAST_MSG "Creating external docker network"
  DOCKER_NETWORK_SETUP $DOCKER_HOST_ARG

  BROADCAST_MSG "Removing containers and cleaning copied folders"
  NETWORK_DOWN $DOCKER_HOST_ARG

  CALL_PREPARE_CHAINCODES

  BROADCAST_MSG "chaincode files are picked up from $CC_BUILD_DIR"
  BROADCAST_MSG "COMPOSE_DIR is $COMPOSE_DIR"

  BROADCAST_MSG "STARTING UP COMPOSE"
  bash -c "$DOCKER_HOST_ARG docker-compose ${COMPOSE_OPTS} up -d"

  BROADCAST_MSG "Tailing fabric-tools"
  bash -c "$DOCKER_HOST_ARG docker logs fabric-tools -f"

  BROADCAST_MSG "Removing Copied files"
  REMOVE_COPIED_FILES
}

export UP_DOWN=$1

VALIDATE_ARGS

REMOTE_HOST=$2
if [ "$REMOTE_HOST" ] ; then
    DOCKER_HOST_ARG="DOCKER_HOST=$REMOTE_HOST"
fi

echo COMPOSE_DIR is $COMPOSE_DIR

if [ -z "$COMPOSE_DIR" ] ; then
  export COMPOSE_DIR=${GOPATH}/src/github.com/securekey/marbles-perf/deployment/fabric
fi

if [ "${UP_DOWN}" == "up" ]; then
	NETWORK_UP $DOCKER_HOST_ARG
elif [ "${UP_DOWN}" == "down" ]; then
	NETWORK_DOWN $DOCKER_HOST_ARG
elif [ "${UP_DOWN}" == "restart" ]; then
	NETWORK_DOWN $DOCKER_HOST_ARG
	NETWORK_UP $DOCKER_HOST_ARG
else
	PRINT_HELP
	exit 1
fi
