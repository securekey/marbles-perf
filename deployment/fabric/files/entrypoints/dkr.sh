#!/bin/sh
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
wait_for_peer() {
    peer_idx=$1

    peer_PID=$(docker exec -it fabric-peer$peer_idx pgrep peer)
    while ! ( echo "$peer_PID" | grep -q "$EXPECTED_PID" ) ; do
        peer_PID=$(docker exec -it fabric-peer$peer_idx pgrep peer)
        echo "waiting for fabric-peer$peer_idx to become main process"
        sleep 2
    done
    echo "fabric-peer$peer_idx is running as main process."
    touch /tmp/deploy-status/started_peer$peer_idx
}

if [ "$ARCH" = 's390x' ]; then
    echo "installing pgrep for all fabric containers because platform is \"$ARCH\""
    docker exec -it fabric-ca /bin/sh -c 'apt-get update'
    docker exec -it fabric-ca /bin/sh -c 'apt-get install -y procps'

    docker exec -it fabric-ca2 /bin/sh -c 'apt-get update'
    docker exec -it fabric-ca2 /bin/sh -c 'apt-get install -y procps'

    docker exec -it fabric-ca-sk /bin/sh -c 'apt-get update'
    docker exec -it fabric-ca-sk /bin/sh -c 'apt-get install -y procps'

    docker exec -it fabric-caOrderer /bin/sh -c 'apt-get update'
    docker exec -it fabric-caOrderer /bin/sh -c 'apt-get install -y procps'

    docker exec -it fabric-orderer0 /bin/sh -c 'apt-get update'
    docker exec -it fabric-orderer0 /bin/sh -c 'apt-get install -y procps'

    docker exec -it fabric-peer1 /bin/sh -c 'apt-get update'
    docker exec -it fabric-peer1 /bin/sh -c 'apt-get install -y procps'

    docker exec -it fabric-peer0 /bin/sh -c 'apt-get update'
    docker exec -it fabric-peer0 /bin/sh -c 'apt-get install -y procps'

    docker exec -it fabric-peer1b /bin/sh -c 'apt-get update'
    docker exec -it fabric-peer1b /bin/sh -c 'apt-get install -y procps'

    docker exec -it fabric-peer0b /bin/sh -c 'apt-get update'
    docker exec -it fabric-peer0b /bin/sh -c 'apt-get install -y procps'

    docker exec -it fabric-peer9 /bin/sh -c 'apt-get update'
    docker exec -it fabric-peer9 /bin/sh -c 'apt-get install -y procps'
fi

EXPECTED_PID="^1"
CA_PID=$(docker exec -it fabric-ca pgrep fabric-ca-serve)

while ! ( echo "$CA_PID" | grep -q "$EXPECTED_PID" ) ; do
    CA_PID=$(docker exec -it fabric-ca pgrep fabric-ca-serve)
    echo "waiting for fabric-ca to become main process"
    sleep 2
done
echo "fabric-ca is running as main process."
touch /tmp/deploy-status/fabric-ca-ready

CA2_PID=$(docker exec -it fabric-ca2 pgrep fabric-ca-serve)
while ! ( echo "$CA2_PID" | grep -q "$EXPECTED_PID" ) ; do
    CA2_PID=$(docker exec -it fabric-ca2 pgrep fabric-ca-serve)
    echo "waiting for fabric-ca2 to become main process"
    sleep 2
done
echo "fabric-ca2 is running as main process."
touch /tmp/deploy-status/fabric-ca2-ready

CA_SK_PID=$(docker exec -it fabric-ca-sk pgrep fabric-ca-serve)
while ! ( echo "$CA_SK_PID" | grep -q "$EXPECTED_PID" ) ; do
    CA_SK_PID=$(docker exec -it fabric-ca-sk pgrep fabric-ca-serve)
    echo "waiting for fabric-ca-sk to become main process"
    sleep 2
done
echo "fabric-ca-sk is running as main process."
touch /tmp/deploy-status/fabric-ca-sk-ready

CA_ORDERER_PID=$(docker exec -it fabric-caOrderer pgrep fabric-ca-serve)
while ! ( echo "$CA_ORDERER_PID" | grep -q "$EXPECTED_PID" ) ; do
    CA_ORDERER_PID=$(docker exec -it fabric-caOrderer pgrep fabric-ca-serve)
    echo "waiting for fabric-caOrderer to become main process"
    sleep 2
done
echo "fabric-caOrderer is running as main process."
touch /tmp/deploy-status/fabric-caOrderer-ready


CA_CLIENT_STATUS=$(docker ps -a --filter name=^/fabric-ca-client$ --format "{{.Status}}")
while ! ( echo "$CA_CLIENT_STATUS" | grep -q "Exited (0)" ) ; do
    CA_CLIENT_STATUS=$(docker ps -a --filter name=^/fabric-ca-client$ --format "{{.Status}}")
    echo "waiting for fabric-ca-client to terminate"
    sleep 2
done
echo "fabric-ca-client has finished registering."

CA_CLIENT2_STATUS=$(docker ps -a --filter name=^/fabric-ca-client2$ --format "{{.Status}}")
while ! ( echo "$CA_CLIENT2_STATUS" | grep -q "Exited (0)" ) ; do
    CA_CLIENT2_STATUS=$(docker ps -a --filter name=^/fabric-ca-client2$ --format "{{.Status}}")
    echo "waiting for fabric-ca-client2 to terminate"
    sleep 2
done
echo "fabric-ca-client2 has finished registering."

CA_CLIENT_SK_STATUS=$(docker ps -a --filter name=^/fabric-ca-client-sk$ --format "{{.Status}}")
while ! ( echo "$CA_CLIENT_SK_STATUS" | grep -q "Exited (0)" ) ; do
    CA_CLIENT_SK_STATUS=$(docker ps -a --filter name=^/fabric-ca-client-sk$ --format "{{.Status}}")
    echo "waiting for fabric-ca-client-sk to terminate"
    sleep 2
done
echo "fabric-ca-client-sk has finished registering."

CA_ORDERER_CLIENT_STATUS=$(docker ps -a --filter name=^/fabric-ca-clientOrderer$ --format "{{.Status}}")
while ! ( echo "$CA_ORDERER_CLIENT_STATUS" | grep -q "Exited (0)" ) ; do
    CA_ORDERER_CLIENT_STATUS=$(docker ps -a --filter name=^/fabric-ca-clientOrderer$ --format "{{.Status}}")
    echo "waiting for fabric-ca-clientOrderer to terminate"
    sleep 2
done
echo "fabric-ca-clientOrderer has finished registering."
touch /tmp/deploy-status/done_ca_registration

ORDERER_PID=$(docker exec -it fabric-orderer0 pgrep orderer)
while ! ( echo "$ORDERER_PID" | grep -q "$EXPECTED_PID" ) ; do
    ORDERER_PID=$(docker exec -it fabric-orderer0 pgrep orderer)
    echo "waiting for fabric-orderer to become main process"
    sleep 2
done
echo "fabric-orderer0 is running as main process."
touch /tmp/deploy-status/started_orderer

for peer in 0 1 9 0b 1b 9b 0c 1c 9c ; do
    wait_for_peer $peer
done
