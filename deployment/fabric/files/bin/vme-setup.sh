#!/bin/bash

PEERS_MYBANK1="peer0 peer0b peer0c"
PEERS_MYBANK2="peer1 peer1b peer1c"
PEERS_SECUREKEY="peer9 peer9b peer9c"
PEERS_CONSORTIUM="${PEERS_MYBANK1} ${PEERS_MYBANK2} ${PEERS_SECUREKEY}"

ADMIN_USER_ID=${ADMIN_USER_ID:-admin}


function VERIFY_RESULT {
	if [ $1 -ne 0 ] ; then
		echo "!!!!!!!!!!!!!!! "$2" !!!!!!!!!!!!!!!!"
		BROADCAST "FATAL ERROR - EXITING"
		echo
		exit 1
	fi
}

function BROADCAST {
	local MESSAGE=$1

	echo
	echo "======================================================================="
	echo "   > ${MESSAGE}"
	echo "======================================================================="
	echo
}

function BROADCAST_RESULT {
	local MESSAGE=$1

	echo "-----------------------------------------------------------------------"
	echo "  ${MESSAGE}"
	echo "-----------------------------------------------------------------------"
}

function SET_PEER_ENV {
	local PEER_NAME=${1}
	local LOG_LEVEL=${2:-DEBUG}
    local PEER_MSP_ID="Org1MSP"
    local ADMIN_MSP_DIR=/data/adminOrg1MSP

    peer_prefix=${PEER_NAME:0:5}

  case ${peer_prefix} in
  peer0)
    PEER_MSP_ID="mybank1"
    ADMIN_MSP_DIR=/data/msp_admin_mybank1
    ;;
  peer1)
    PEER_MSP_ID="mybank2"
    ADMIN_MSP_DIR=/data/msp_admin_mybank2
    ;;
  peer9)
    PEER_MSP_ID="securekey"
    ADMIN_MSP_DIR=/data/msp_admin_securekey
    ;;
  *)
    PEER_MSP_ID="Org1MSP"
    ADMIN_MSP_DIR=/data/adminOrg1MSP
    ;;
  esac

	BROADCAST "SETTING TARGET PEER: ${PEER_NAME} MSP_ID=${PEER_MSP_ID}"

	export CORE_LOGGING_LEVEL=${LOG_LEVEL}
	export CORE_PEER_LOCALMSPID=${PEER_MSP_ID}
	export CORE_PEER_TLS_ENABLED=true
	export CORE_PEER_TLS_ROOTCERT_FILE=/data/tls/ca_root.pem
	export CORE_PEER_MSPCONFIGPATH=${ADMIN_MSP_DIR}
	export CORE_PEER_ADDRESS=${PEER_NAME}.${PEER_DOMAIN}:7051

  # MUTUAL TLS
  export CORE_PEER_TLS_CLIENTAUTHREQUIRED=true
  export CORE_PEER_TLS_CLIENTROOTCAS_FILES=/data/tls/ca_root.pem
  export CORE_PEER_TLS_CLIENTCERT_FILE=/data/tls/server_client_wild_vme_sk_dev.pem
  export CORE_PEER_TLS_CLIENTKEY_FILE=/data/tls/server_client_wild_vme_sk_dev-key.pem

	if [ "${LOG_LEVEL}" = "DEBUG"  ]; then
		BROADCAST "SETTING PEER ENV"
		env | grep CORE
	fi

    # symlink admin cert to a location and name pattern that fabric-sdk can use for configcli
    # the below naming pattern assumes OrgID == MSPId
    #
    admin_cert_file=${CORE_PEER_MSPCONFIGPATH}/admincerts/cert.pem
    sdk_cert_file=${CORE_PEER_MSPCONFIGPATH}/signcerts/admin\@${PEER_MSP_ID}-cert.pem

    if [ -r ${admin_cert_file} ] && [ ! -r ${sdk_cert_file} ] ; then
        ln -s ${admin_cert_file} ${sdk_cert_file}
    fi
}

function CREATE_CHANNEL {
	PEER_NAME=${1}
	local CHANNEL_NAME=$2

	SET_PEER_ENV ${PEER_NAME}
	BROADCAST "CREATING CHANNEL: ${CHANNEL_NAME}"

	$PEER_BIN channel create -o ${ORDERER_HOST}:7050 -c ${CHANNEL_NAME} \
	-f ${CHANNEL_BASEDIR}/${CHANNEL_NAME}.tx --tls ${CORE_PEER_TLS_ENABLED} --cafile ${ORDERER_CA_CERT} >&log.txt

	res=$?
	cat log.txt
	# DISABLE FOR NOW, WITH KAFKA ENABLED YOU GET AN ERROR BUT IT ACTUALLY WORKS
	VERIFY_RESULT $res "Channel creation failed"
	BROADCAST_RESULT "Channel ${CHANNEL_NAME} was created successfully"
	sleep 2
}

function JOIN_CHANNEL {
	PEER_NAME=${1}
	local CHANNEL_NAME=$2

	SET_PEER_ENV ${PEER_NAME}
	BROADCAST "${CORE_PEER_ADDRESS} JOINING CHANNEL: ${CHANNEL_NAME}"

	$PEER_BIN channel join -b ${CHANNEL_NAME}.block  >&log.txt

	res=$?
	cat log.txt
	if [ $res -ne 0 -a $COUNTER -lt $MAX_RETRY ]; then
		COUNTER=` expr $COUNTER + 1`
		BROADCAST_RESULT "${CORE_PEER_ADDRESS} failed to JOIN. Retrying.."
		sleep 1
		JOIN_CHANNEL ${PEER_NAME} ${CHANNEL_NAME}
	else
		COUNTER=0
	fi

	VERIFY_RESULT $res "After $MAX_RETRY attempts, ${CORE_PEER_ADDRESS} has failed to Join the Channel"
	sleep 1
}

function PACKAGE_CHAINCODE {
	PEER_NAME=${1}
	local CC_NAME=$2
	local CC_VERSION=$3
	local CC_PATH=$4

	SET_PEER_ENV ${PEER_NAME}
	cd ${PACKAGES_DIR}
	BROADCAST "${CORE_PEER_ADDRESS} PACKAGING CHAINCODE: ${CC_NAME}"

	# signing package does not seem to install. disable for now: -s -S -i "AND('Org1.admin')"
	$PEER_BIN chaincode package \
	-n ${CC_NAME} -v ${CC_VERSION} \
	-p ${CC_PATH} \
	${CC_NAME}_v${CC_VERSION}.out >&log.txt

	res=$?
	cat log.txt
	VERIFY_RESULT $res "Chaincode packaging of ${CC_NAME} on remote peer ${CORE_PEER_ADDRESS} has Failed"
	BROADCAST_RESULT "Chaincode package ${PWD}/${CC_NAME}_v${CC_VERSION}.out was successfully created."
	sleep 1

}

function UPDATE_ANCHORPEERS {
	PEER_NAME=${1}
	local CHANNEL_NAME=$2
	local CHANNEL_TX=$3

  BROADCAST "UPDATE ANCHORPEERS: ${CHANNEL_NAME}"
	SET_PEER_ENV ${PEER_NAME}

	$PEER_BIN channel update -o ${ORDERER_HOST}:7050 \
	-c ${CHANNEL_NAME} -f ${CHANNEL_BASEDIR}/${CHANNEL_TX}.tx \
	--tls $CORE_PEER_TLS_ENABLED --cafile ${ORDERER_CA_CERT} >&log.txt

	res=$?
	cat log.txt
	VERIFY_RESULT $res "Update Anchor Peers on remote peer ${CORE_PEER_ADDRESS} has Failed"
	BROADCAST_RESULT "Update Anchor Peers Successful"
	sleep 1

}


function INSTALL_CHAINCODE_PACKAGE {
	PEER_NAME=${1}
	local CC_PKG=$2

	SET_PEER_ENV ${PEER_NAME}
	BROADCAST "${CORE_PEER_ADDRESS} INSTALLING CHAINCODE PKG: ${CC_PKG}"

	$PEER_BIN chaincode install \
	${CC_PKG} >&log.txt

	res=$?
	cat log.txt
	VERIFY_RESULT $res "Chaincode package installation on remote peer ${CORE_PEER_ADDRESS} has Failed"
	BROADCAST_RESULT "Chaincode package ${CC_PKG} is installed successfully on remote peer ${CORE_PEER_ADDRESS}"
	sleep 1

}

function INSTALL_CHAINCODE {
	PEER_NAME=${1}
	local CC_NAME=$2
	local CC_VERSION=$3
	local CC_PATH=$4

	SET_PEER_ENV ${PEER_NAME}
	BROADCAST "${CORE_PEER_ADDRESS} INSTALLING CHAINCODE: ${CC_NAME}"

	$PEER_BIN chaincode install \
	-n ${CC_NAME} -v ${CC_VERSION} \
	-p ${CC_PATH} >&log.txt

	res=$?
	cat log.txt
	VERIFY_RESULT $res "Chaincode installation on remote peer ${CORE_PEER_ADDRESS} has Failed"
	BROADCAST_RESULT "Chaincode ${CC_NAME} is installed successfully on remote peer ${CORE_PEER_ADDRESS}"
	sleep 1

}

function INIT_CHAINCODE_WITH_POLICY {
	PEER_NAME=${1}
	local CHANNEL_NAME=$2
	local CC_NAME=$3
	local CC_VERSION=$4
	local CC_POLCIY=$5
	local CC_CONSTRUCTOR=$6
  local CC_COLPOLICY=$7

	SET_PEER_ENV ${PEER_NAME}
	BROADCAST "${CORE_PEER_ADDRESS} INIT CHAINCODE WITH POLICY: ${CC_NAME}"


    if [ ! -z "$CC_COLPOLICY" ] ; then
        COLLECTION_CONFIG="--collections-config ${CC_COLPOLICY}"
    fi
	$PEER_BIN chaincode instantiate -o ${ORDERER_HOST}:7050 \
      ${COLLECTION_CONFIG} \
	  --tls ${CORE_PEER_TLS_ENABLED} --cafile ${ORDERER_CA_CERT} -C ${CHANNEL_NAME} \
	  -n ${CC_NAME} -v ${CC_VERSION} -c "${CC_CONSTRUCTOR}" -P "${CC_POLCIY}" >&log.txt

	res=$?
	cat log.txt
	VERIFY_RESULT $res "Chaincode ${CC_NAME} instantiation on ${CORE_PEER_ADDRESS} on channel '${CHANNEL_NAME}' failed"
	BROADCAST_RESULT "Chaincode ${CC_NAME} Instantiation on ${CORE_PEER_ADDRESS} on channel '$CHANNEL_NAME' is successful"
	sleep 1

}

function INIT_CHAINCODE {
	PEER_NAME=${1}
	local CHANNEL_NAME=$2
	local CC_NAME=$3
	local CC_VERSION=$4
	local CC_POLCIY=$5
	local CC_CONSTRUCTOR=$6

	SET_PEER_ENV ${PEER_NAME}
	BROADCAST "${CORE_PEER_ADDRESS} INIT CHAINCODE: ${CC_NAME}"

	$PEER_BIN chaincode instantiate -o ${ORDERER_HOST}:7050 \
	--tls ${CORE_PEER_TLS_ENABLED} --cafile ${ORDERER_CA_CERT} -C ${CHANNEL_NAME} \
	-n ${CC_NAME} -v ${CC_VERSION} -c "${CC_CONSTRUCTOR}" -P "${CC_POLCIY}" >&log.txt

	res=$?
	cat log.txt
	VERIFY_RESULT $res "Chaincode ${CC_NAME} instantiation on ${CORE_PEER_ADDRESS} on channel '${CHANNEL_NAME}' failed"
	BROADCAST_RESULT "Chaincode ${CC_NAME} Instantiation on ${CORE_PEER_ADDRESS} on channel '$CHANNEL_NAME' is successful"
	sleep 1

}


function INVOKE_CHAINCODE {
	PEER_NAME=${1}
	local CHANNEL_NAME=$2
	local CC_NAME=$3
	local CC_VERSION=$4
	local CC_CONSTRUCTOR=$5

	SET_PEER_ENV ${PEER_NAME}
	BROADCAST "${CORE_PEER_ADDRESS} INVOKE CHAINCODE: ${CC_NAME}"

	$PEER_BIN chaincode invoke -o ${ORDERER_HOST}:7050 \
	--tls ${CORE_PEER_TLS_ENABLED} --cafile ${ORDERER_CA_CERT} -C ${CHANNEL_NAME} \
	-n ${CC_NAME} -v ${CC_VERSION} -c "${CC_CONSTRUCTOR}" >&log.txt

	res=$?
	cat log.txt
	VERIFY_RESULT $res "Chaincode invoke for ${CC_NAME} on ${CORE_PEER_ADDRESS} on channel '${CHANNEL_NAME}' failed"
	BROADCAST_RESULT "Chaincode invoke for ${CC_NAME} on ${CORE_PEER_ADDRESS} on channel '$CHANNEL_NAME' is successful"
	sleep 0.5

}

# cc warmup no longer needed since fabric 1.2 update, keeping code for future use
# ignore output
function WARMUP_QUERY_IGNORE_ERROR {
  PEER_NAME=${1}
	local CHANNEL_NAME=$2
	local CC_NAME=$3
	local CC_QUERY=$4

	SET_PEER_ENV ${PEER_NAME}
	BROADCAST "${CORE_PEER_ADDRESS} WARMUP CHAINCODE IGNORE ERROR: ${CC_NAME}"
	
	$PEER_BIN chaincode query \
	--tls ${CORE_PEER_TLS_ENABLED} -C ${CHANNEL_NAME} \
	-n ${CC_NAME} -c "${CC_QUERY}" >& /dev/null
}

function WARMUP_QUERY {
	PEER_NAME=${1}
	local CHANNEL_NAME=$2
	local CC_NAME=$3
	local CC_QUERY=$4

	SET_PEER_ENV ${PEER_NAME}
	BROADCAST "${CORE_PEER_ADDRESS} WARMUP CHAINCODE: ${CC_NAME}"

	$PEER_BIN chaincode query \
	--tls ${CORE_PEER_TLS_ENABLED} -C ${CHANNEL_NAME} \
	-n ${CC_NAME} -c "${CC_QUERY}" >&log.txt

	res=$?
	cat log.txt
	VERIFY_RESULT $res "Chaincode warmup queried on ${CORE_PEER_ADDRESS} on channel '${CHANNEL_NAME}' failed"
	BROADCAST_RESULT "Chaincode warmup queried on ${CORE_PEER_ADDRESS} on channel '${CHANNEL_NAME}' successful"
	#sleep 1
}

function QUERY_CHAINCODE {
	PEER_NAME=${1}
	local CHANNEL_NAME=$2
	local CC_NAME=$3
	local CC_QUERY=$4

	SET_PEER_ENV ${PEER_NAME}
	BROADCAST "${CORE_PEER_ADDRESS} QUERY CHAINCODE: ${CC_NAME}"

	$PEER_BIN chaincode query \
	--tls ${CORE_PEER_TLS_ENABLED} -C ${CHANNEL_NAME} \
	-n ${CC_NAME} -v ${CC_VERSION} -c "${CC_QUERY}" >&log.txt

	res=$?
	cat log.txt
	VERIFY_RESULT $res "Chaincode Queried on ${CORE_PEER_ADDRESS} on channel '${CHANNEL_NAME}' failed"
	BROADCAST_RESULT "Chaincode Queried on ${CORE_PEER_ADDRESS} on channel '${CHANNEL_NAME}' is successful"
	#sleep 1
}

function FETCH_BLOCK {
	local CHANNEL_NAME=$1
	local BLOCK_TO_FETCH=$2

	BROADCAST "FETCHING BLOCK ${BLOCK_TO_FETCH} FROM CHANNEL ${CHANNEL_NAME}"

	# when using TLS the block is always called true
	$PEER_BIN channel fetch ${BLOCK_TO_FETCH} \
	--channelID ${CHANNEL_NAME} \
	-o ${ORDERER_HOST}:7050 \
	--tls ${CORE_PEER_TLS_ENABLED} \
	--cafile ${ORDERER_CA_CERT}

	VERIFY_RESULT $? "Failed to fetch block"
	mv true ${CHANNEL_NAME}_${BLOCK_TO_FETCH}.block
	BROADCAST_RESULT "The block was fetched successfully"

}

function VIEW_BLOCK {
	local BLOCK_FILE=$1
	local PROFILE=$2
	local CHANNEL=$3

	BROADCAST "VIEWING BLOCK ${BLOCK_FILE}"

	FABRIC_CFG_PATH=/data \
	configtxgen -profile ${PROFILE} \
	-channelID ${CHANNEL} \
	-inspectBlock ${BLOCK_FILE}

}



################################################################################
# START Setup...
################################################################################

if [ ! -z "${WAIT_SIGNAL_FILES}" ] ; then
	wait-for-files.sh ${WAIT_SIGNAL_FILES}
fi

PEER_BIN=/usr/local/bin/peer
CHANNEL_BASEDIR=/data/channel-artifacts
export PACKAGES_DIR=${GOPATH}/src/gerrit.securekey.com/user-cc

# COMMON ENV SPECS -------------------------------------------------------------
ORDERER_HOST="orderer0.vme.sk.dev"
ORDERER_CA_CERT=/data/tls/ca_root.pem

PEER_PREFIX=""
PEER_DOMAIN="vme.sk.dev"

COUNTER=0
MAX_RETRY=5

if [ ! -z "${CC_TO_UPGRADE}" ] ; then
	echo "Will upgrade chaincodes: ${CC_TO_UPGRADE}"
	for cc in ${CC_TO_UPGRADE//,/ }; do
		UPGRADE_CC ${cc}
	done
	exit 0
fi

cd /data/channel-artifacts
/data/create_channel_tx.sh

# CHANNELS ---------------------------------------------------------------------
# CREATE CHANNEL consortium; all peers will join
CREATE_CHANNEL ${PEERS_CONSORTIUM/%\ */} consortium
UPDATE_ANCHORPEERS ${PEERS_MYBANK1/%\ */} consortium anchors-mybank1_consortium
UPDATE_ANCHORPEERS ${PEERS_MYBANK2/%\ */} consortium anchors-mybank2_consortium
UPDATE_ANCHORPEERS ${PEERS_SECUREKEY/%\ */} consortium anchors-securekey_consortium

for p in ${PEERS_CONSORTIUM}; do
	JOIN_CHANNEL ${p} consortium
done

# DISPLAY CONFIG BLOCKS FOR CHANNELS
SET_PEER_ENV peer0
FETCH_BLOCK consortium config
VIEW_BLOCK consortium_config.block VmeConsortiumChannel consortium

# SNAP CONFIGURATIONS ----------------------------------------------------------
#cp -p /opt/securekey/libs/configcli /usr/local/bin/



# CHAINCODE --------------------------------------------------------------------
# Package chaincode
#for cc in fmp txaudit consortium custodian dlbp steward; do
for cc in marbles; do
	PACKAGE_CHAINCODE ${PEERS_CONSORTIUM/%\ */} ${cc}cc "1.0" gerrit.securekey.com/user-cc/chaincodes/${cc}_cc
done


# INSTALL CHAINCODE
for p in ${PEERS_CONSORTIUM}; do
#	for cc in fmp txaudit consortium custodian dlbp steward; do
	for cc in marbles; do
		INSTALL_CHAINCODE_PACKAGE ${p} ${PACKAGES_DIR}/${cc}cc_v1.0.out
	done
done

sleep 2


CC_POLICY_CONSORTIUM="OutOf(2, 'mybank1.member', 'mybank2.member')"
CC_POLICY_DEFAULT="OR('mybank1.member', 'mybank2.member', 'securekey.member')"
# Instantiating chaincodes on consortium
INIT_CHAINCODE ${PEERS_CONSORTIUM/%\ */} consortium marblescc "1.0" "$CC_POLICY_DEFAULT" '{"Args":[]}'


# INVOKES ----------------------------------------------------------------------

# wait for fmpcc install complete...
BROADCAST "sleeping for 3 seconds to allow for marbles init"
sleep 3

# need to do some warm ups. these calls should fail
BROADCAST "DOING WARMUPS"

for p in ${PEERS_CONSORTIUM}; do
  for cc in marbles; do
    WARMUP_QUERY_IGNORE_ERROR ${p} consortium ${cc}cc '{"Args":["warmup"]}'
  done
done

BROADCAST "SCRIPT HAS COMPLETED!"

if [ ! -z "${WRITE_SIGNAL_FILE}" ] ; then
  touch "${WRITE_SIGNAL_FILE}"
fi

# Leave container running if DEBUG_ENABLED is true
if [ ${DEBUG_ENABLED:-false} == true ]; then
  BROADCAST "DEBUG_ENABLED set to true. Container will NOT exit!"
  tail -f /dev/null
fi

