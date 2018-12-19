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

FABRIC_CA_FQDN="${1:-localhost}"

REGISTER_USER() {
  local USER_ID=$1
  local USER_TYPE=$2
  local USER_AFF="${3:-org1}"

  fabric-ca-client register \
    --id.name ${USER_ID} \
    --id.secret testing \
    --id.type ${USER_TYPE} \
    --id.affiliation ${USER_AFF} \
    --url https://${FABRIC_CA_FQDN}:7054

  sleep 0.25
}

echo ">>>> ENROLLING ADMIN <<<<"
fabric-ca-client enroll --url https://admin:adminpw@${FABRIC_CA_FQDN}:7054

cp -rp /etc/hyperledger/fabric-ca-server/msp/signcerts /etc/hyperledger/fabric-ca-server/msp/admincerts

REGISTER_USER adapter.mybank.com client
REGISTER_USER adapter.mybank2.com client
REGISTER_USER adapter.securekey.com client

REGISTER_USER peer0 peer
REGISTER_USER peer0b peer
REGISTER_USER peer0c peer
REGISTER_USER peer1 peer
REGISTER_USER peer1b peer
REGISTER_USER peer1c peer
REGISTER_USER peer9 peer
REGISTER_USER peer9b peer
REGISTER_USER peer9c peer
REGISTER_USER orderer0 orderer


