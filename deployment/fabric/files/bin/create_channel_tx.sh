# generate channel tx for consortium
FABRIC_CFG_PATH=/data \
 configtxgen -profile VmeConsortiumChannel \
 -channelID consortium \
 -outputCreateChannelTx /data/channel-artifacts/consortium.tx \
 -inspectChannelCreateTx /data/channel-artifacts/consortium.tx

# generate securekey anchor peer tx for consortium
FABRIC_CFG_PATH=/data \
 configtxgen -profile VmeConsortiumChannel \
 -channelID consortium \
 -outputAnchorPeersUpdate /data/channel-artifacts/anchors-securekey_consortium.tx \
 -asOrg securekey

# generate mybank1 anchor peer tx for consortium
FABRIC_CFG_PATH=/data \
 configtxgen -profile VmeConsortiumChannel \
 -channelID consortium \
 -outputAnchorPeersUpdate /data/channel-artifacts/anchors-mybank1_consortium.tx \
 -asOrg mybank1

# generate mybank2 anchor peer tx for consortium
FABRIC_CFG_PATH=/data \
 configtxgen -profile VmeConsortiumChannel \
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
