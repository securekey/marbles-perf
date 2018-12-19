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


#
#  FUNCTIONS
#

# Ensure that the following environment variables are set
function CHECK_PARAMS {

    if [ -z "${FABRIC_CHANNEL}" ]; then
        echo "ERROR: FABRIC_CHANNEL is not specified!"; exit 1
    fi

    if [ -z "${FABRIC_MSPID}" ]; then
        echo "ERROR: FABRIC_MSPID is not specified!"; exit 1
    fi

    if [ -z "${FABRIC_PEER1_URL}" ]; then
        echo "ERROR: FABRIC_PEER1_URL is not specified!"; exit 1
    fi

    if [ -z "${FABRIC_PEER2_URL}" ]; then
        echo "ERROR: FABRIC_PEER2_URL is not specified!"; exit 1
    fi

    if [ -z "${FABRIC_SKPEER1_URL}" ]; then
        echo "ERROR: FABRIC_SKPEER1_URL is not specified!"; exit 1
    fi

    if [ -z "${FABRIC_SKPEER2_URL}" ]; then
        echo "ERROR: FABRIC_SKPEER2_URL is not specified!"; exit 1
    fi

    if [ -z "${FABRIC_TLS_CA_CERTS}" ]; then
        echo "ERROR: FABRIC_TLS_CA_CERTS is not specified!"; exit 1
    fi
}


# Function deployFabricTLS is looking for environmental variables
# related to ca certificates and client certificate / key for
# mutual tls to be enbaled for Hyperledger Fabric
function deployFabricTLS {

    if [ "x${FABRIC_TLS_CLIENT_CERT_BASE64}" != "x" ]; then

        # generating the client certificate
        echo "Processing FABRIC_TLS_CLIENT_CERT_BASE64 ..."
        echo "${FABRIC_TLS_CLIENT_CERT_BASE64}" | base64 -d > ${SUPPORT_TOOL_HOME}/_client_cert.pem

        if [ "x$?" != "x0" ]; then
            echo "ERROR: deployFabricTLS unable to base64 decode 'FABRIC_TLS_CLIENT_CERT_BASE64'"
            exit 1
        fi

        # if it looks like a certificate overwrite the
        # environmental variable of the adapter
        header=$(head -1 ${SUPPORT_TOOL_HOME}/_client_cert.pem)

        if [ "x${header}" == "x-----BEGIN CERTIFICATE-----" ]; then

            export FABRIC_TLS_CLIENT_CERT=${SUPPORT_TOOL_HOME}/_client_cert.pem
            echo "FABRIC_TLS_CLIENT_CERT set to ${FABRIC_TLS_CLIENT_CERT}"
        else
            echo "Valid certificate header not found ..."
        fi
    fi

    if [ "x${FABRIC_TLS_CLIENT_KEY_BASE64}" != "x" ]; then

        # generating the client key
        echo "Processing FABRIC_TLS_CLIENT_KEY_BASE64 ..."
        echo "${FABRIC_TLS_CLIENT_KEY_BASE64}" | base64 -d > ${SUPPORT_TOOL_HOME}/_client_key.pem

        if [ "x$?" != "x0" ]; then
            echo "ERROR: deployFabricTLS unable to base64 decode 'FABRIC_TLS_CLIENT_KEY_BASE64'"
            exit 1
        fi

        # if it looks like a key overwrite the
        # environmental variable of the adapter
        header=$(head -1 ${SUPPORT_TOOL_HOME}/_client_key.pem)

        if [ "x${header}" == "x-----BEGIN EC PRIVATE KEY-----" ] || [ "x${header}" == "x-----BEGIN PRIVATE KEY-----" ]; then

            export FABRIC_TLS_CLIENT_KEY=${SUPPORT_TOOL_HOME}/_client_key.pem
            echo "FABRIC_TLS_CLIENT_KEY set to ${FABRIC_TLS_CLIENT_KEY}"
        else
            echo "Valid key header not found ..."
        fi
    fi

    if [ "x${FABRIC_CA_CERTS_BASE64}" != "x" ]; then

        # generating ca certificates
        echo "Processing FABRIC_CA_CERTS_BASE64 ..."
        echo "${FABRIC_CA_CERTS_BASE64}" | base64 -d > ${SUPPORT_TOOL_HOME}/_ca_certs.pem

        if [ "x$?" != "x0" ]; then
            echo "ERROR: deployFabricTLS unable to base64 decode 'FABRIC_CA_CERTS_BASE64'"
            exit 1
        fi

        # this file can have multiple ca certs, but if the first line looks
        # correct overwrite the environmental variable of the adapter
        header=$(head -1 ${SUPPORT_TOOL_HOME}/_ca_certs.pem)

        if [ "x${header}" == "x-----BEGIN CERTIFICATE-----" ]; then

            export FABRIC_TLS_CA_CERTS=${SUPPORT_TOOL_HOME}/_ca_certs.pem
            echo "FABRIC_TLS_CA_CERTS set to ${FABRIC_TLS_CA_CERTS}"
        else
            echo "Valid certificate header not found ..."
        fi
    fi
}

# Usage: decodeCertificateList <certchain_file> <destination_folder>
#
# Function decodeCertificateList takes a concatenated list of certificates
# and creates individual files out of them in a specified folder. If the folder
# does not exist it will make an attempt to create it.
function decodeCertificateList() {

    certchain_file=${1}
    dest_folder=${2}

    # missing parameter?
    if [ "x${certchain_file}" == "x" ] || [ "x${dest_folder}" == "x" ]; then

        echo "ERROR: decodeCertificateList missing parameter(s)"
        exit 1
    fi

    # file does not exist
    if [ ! -f "${certchain_file}" ]; then

        echo "ERROR: decodeCertificateList file '${certchain_file}' not found"
        exit 1
    fi

    # the folder need to be created?
    if [ ! -d "${dest_folder}" ]; then

        mkdir -p "${dest_folder}"

        if [ "x$?" != "x0" ]; then
            echo "ERROR: decodeCertificateList unable to create directory '${dest_folder}'"
            exit 1
        fi
    fi

    # certificate chain splitting
    cat  "${certchain_file}" | awk -v dest=${dest_folder} '{print > dest"/cert" (1+n) ".pem"} /-----END CERTIFICATE-----/ {n++}'
}

# Function updateOSCertificateTruststore is looking for an environment variable called
# OS_TRUSTSTORE_CERT_BASE64 and tries to base64 decode it into a file which
# will be picked up by an os trust store update.
function updateOSCertificateTruststore() {

    if [ "x${OS_TRUSTSTORE_CERT_BASE64}" != "x" ]; then

        # extracting the custom ca certificate
        echo "Processing OS_TRUSTSTORE_CERT_BASE64 ..."
        echo "${OS_TRUSTSTORE_CERT_BASE64}" | base64 -d > /tmp/os_trust_certs.pem

        if [ "x$?" != "x0" ]; then

            echo "ERROR: updateOSCertificateTruststore unable to base64 decode 'OS_TRUSTSTORE_CERT_BASE64'"
            exit 1
        fi

        # splitting chain
        decodeCertificateList /tmp/os_trust_certs.pem /etc/pki/ca-trust/source/anchors
    fi

    #
    /usr/bin/update-ca-trust
}

# Function decodeEnrollmentData processes FABRIC_USER_ENROLL_CERT_BASE64 and
# FABRIC_USER_ENROLL_KEY_BASE64 environment variables if any of them defined
function decodeEnrollmentData() {

    if [ "x${FABRIC_USER_ENROLL_CERT_BASE64}" != "x" ]; then

        # check if destination folder is ready
        if [ ! -d "${SUPPORT_TOOL_HOME}/msp/signcerts" ]; then

            # create the directory if it wasn't there
            mkdir -p ${SUPPORT_TOOL_HOME}/msp/signcerts
        fi

        # decoding the enrollment certificate
        echo "Processing FABRIC_USER_ENROLL_CERT_BASE64 ..."

        # calculating the filename
        fn="${SUPPORT_TOOL_HOME}/msp/signcerts/${FABRIC_USER_ID}@${FABRIC_MSPID}-cert.pem"

        echo "Creating ${fn} ..."
        echo "${FABRIC_USER_ENROLL_CERT_BASE64}" | base64 -d > "${fn}"

        if [ "x$?" != "x0" ]; then
            echo "ERROR: decodeEnrollmentData unable to base64 decode 'FABRIC_USER_ENROLL_CERT_BASE64'"
            exit 1
        fi

        # test if the decoded file looks like a certificate
        header=$(head -1 "${fn}")

        if [ "x${header}" != "x-----BEGIN CERTIFICATE-----" ]; then

            echo "Content of the decoded enrollment certificate looks invalid ..."
            exit 1
        fi
    fi

    if [ "x${FABRIC_USER_ENROLL_KEY_BASE64}" != "x" ]; then

        # check if destination folder is ready
        if [ ! -d "${SUPPORT_TOOL_HOME}/msp/keystore" ]; then

            # create the directory if it wasn't there
            mkdir -p ${SUPPORT_TOOL_HOME}/msp/keystore
        fi

        # decoding the enrollment key
        echo "Processing FABRIC_USER_ENROLL_KEY_BASE64 ..."

        # calculating the filename
        fn="${SUPPORT_TOOL_HOME}/msp/keystore/key.pem"

        echo "Creating ${fn} ..."
        echo "${FABRIC_USER_ENROLL_KEY_BASE64}" | base64 -d > "${fn}"

        if [ "x$?" != "x0" ]; then
            echo "ERROR: decodeEnrollmentData unable to base64 decode 'FABRIC_USER_ENROLL_KEY_BASE64'"
            exit 1
        fi

        # test if the decoded key looks like a key
        header=$(head -1 "${fn}")

        if [ "x${header}" != "x-----BEGIN EC PRIVATE KEY-----" ] && [ "x${header}" != "x-----BEGIN PRIVATE KEY-----" ]; then

            echo "Content of the decoded enrollment key looks invalid ..."
            exit 1
        fi
    fi
}

#
#
#

# print some container info
echo "SecureKey Technologies Inc."
echo "---------------------------------------"
echo "Started Marbles Perf server on $(date)"

# Printing user and group ids
echo "uid: $(id -u), gid: $(id -g) ...";

# Printing Version info
echo "Version Information:"
env | grep BUILD_
echo

# If FABRIC_SDK_CLIENT_CONFIGURATION is not defined through an environmental
# variable and a fabric_sdk config file exists, use that
if [ "x${FABRIC_SDK_CLIENT_CONFIGURATION}" == "x" ]; then
    if [ -f "${SUPPORT_TOOL_HOME}/fabric_sdk.yaml" ]; then
        export FABRIC_SDK_CLIENT_CONF_FILE=${SUPPORT_TOOL_HOME}/fabric_sdk.yaml

        # the above sdk requires certain environment variables to be set
        CHECK_PARAMS
    fi
fi

# Looking for missing environmental variables
if [ -z "${FABRIC_ORDERER1_URL}" ]; then
    FABRIC_ORDERER1_URL="grpcs://orderer1:10000"
fi

if [ -z "${FABRIC_ORDERER2_URL}" ]; then
    FABRIC_ORDERER2_URL="grpcs://orderer2:10000"
fi

if [ -z "${FABRIC_ORDERER3_URL}" ]; then
    FABRIC_ORDERER3_URL="grpcs://orderer3:10000"
fi

# Setting hostnames based on URLS provided which is required as of release 10
FABRIC_ORDERER1_CALC=`echo "$FABRIC_ORDERER1_URL" | awk -F/ '{print $3}' | awk -F: '{print $1}'`
FABRIC_ORDERER2_CALC=`echo "$FABRIC_ORDERER2_URL" | awk -F/ '{print $3}' | awk -F: '{print $1}'`
FABRIC_ORDERER3_CALC=`echo "$FABRIC_ORDERER3_URL" | awk -F/ '{print $3}' | awk -F: '{print $1}'`
export FABRIC_ORDERER1=${FABRIC_ORDERER1_CALC:-FABRIC_ORDERER1}
export FABRIC_ORDERER2=${FABRIC_ORDERER2_CALC:-FABRIC_ORDERER2}
export FABRIC_ORDERER3=${FABRIC_ORDERER3_CALC:-FABRIC_ORDERER3}

# Setting peer hostnames which are used for entity matchers
export FABRIC_PEER1_MATCHER=`echo "$FABRIC_PEER1_URL" | awk -F/ '{print $3}'`
export FABRIC_PEER2_MATCHER=`echo "$FABRIC_PEER2_URL" | awk -F/ '{print $3}'`
export FABRIC_SKPEER1_MATCHER=`echo "$FABRIC_SKPEER1_URL" | awk -F/ '{print $3}'`
export FABRIC_SKPEER2_MATCHER=`echo "$FABRIC_SKPEER2_URL" | awk -F/ '{print $3}'`

# If FABRIC_ORDERER_URL is set
# Then use it instead of multiple orderer
if [ -n "${FABRIC_ORDERER_URL}" ]; then
    FABRIC_ORDERER_CALC=`echo "$FABRIC_ORDERER_URL" | awk -F/ '{print $3}' | awk -F: '{print $1}'`
    export FABRIC_ORDERER1="${FABRIC_ORDERER_CALC:-FABRIC_ORDERER1}"
    export FABRIC_ORDERER1_URL="${FABRIC_ORDERER_URL:-FABRIC_ORDERER1_URL}"
fi

# Pickup environmental overwrites from the ENV file. This can be used for
# troubleshooting in extreme situations
if [ -f "${APP_HOME}/ENV" ]; then

    echo "Importing variables from ${SUPPORT_TOOL_HOME}/ENV ..."
    source ${APP_HOME}/ENV
fi

# Starting a shell if specified
if [ "x$@" == "xsh" ] || [ "x$@" == "xbash" ]; then

    echo "Troubleshooting shell opened by the entrypoint."
    echo "Type 'exit' to terminate."
    echo ""

    /bin/sh
    exit 0
fi

echo "";
echo "";

# finally, start the daemon
echo "---"
exec /usr/local/bin/marbles-perf ${APP_CFG}
