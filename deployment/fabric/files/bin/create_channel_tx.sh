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

# generate channel tx for consortium
FABRIC_CFG_PATH=/data \
 configtxgen -profile MPerfConsortiumChannel \
 -channelID consortium \
 -outputCreateChannelTx /data/channel-artifacts/consortium.tx \
 -inspectChannelCreateTx /data/channel-artifacts/consortium.tx

# generate securekey anchor peer tx for consortium
FABRIC_CFG_PATH=/data \
 configtxgen -profile MPerfConsortiumChannel \
 -channelID consortium \
 -outputAnchorPeersUpdate /data/channel-artifacts/anchors-securekey_consortium.tx \
 -asOrg securekey

# generate mybank1 anchor peer tx for consortium
FABRIC_CFG_PATH=/data \
 configtxgen -profile MPerfConsortiumChannel \
 -channelID consortium \
 -outputAnchorPeersUpdate /data/channel-artifacts/anchors-mybank1_consortium.tx \
 -asOrg mybank1

# generate mybank2 anchor peer tx for consortium
FABRIC_CFG_PATH=/data \
 configtxgen -profile MPerfConsortiumChannel \
 -channelID consortium \
 -outputAnchorPeersUpdate /data/channel-artifacts/anchors-mybank2_consortium.tx \
 -asOrg mybank2

# generate channel tx for dlbp-mybank1
FABRIC_CFG_PATH=/data \
  configtxgen -profile DlbpMybank1Channel \
  -channelID dlbp-mybank1 \
  -outputCreateChannelTx /data/channel-artifacts/dlbp-mybank1.tx \
  -inspectChannelCreateTx /data/channel-artifacts/dlbp-mybank1.tx

# generate anchor peer tx for dlbp-mybank1
FABRIC_CFG_PATH=/data \
  configtxgen -profile DlbpMybank1Channel \
  -channelID dlbp-mybank1 \
  -outputAnchorPeersUpdate /data/channel-artifacts/anchors-mybank1-dlbp-mybank1.tx \
  -asOrg mybank1

# generate channel tx for dlbp-mybank2
FABRIC_CFG_PATH=/data \
  configtxgen -profile DlbpMybank2Channel \
  -channelID dlbp-mybank2 \
  -outputCreateChannelTx /data/channel-artifacts/dlbp-mybank2.tx \
  -inspectChannelCreateTx /data/channel-artifacts/dlbp-mybank2.tx

# generate anchor peer tx for dlbp-mybank2
FABRIC_CFG_PATH=/data \
  configtxgen -profile DlbpMybank2Channel \
  -channelID dlbp-mybank2 \
  -outputAnchorPeersUpdate /data/channel-artifacts/anchors-mybank2-dlbp-mybank2.tx \
  -asOrg mybank2
