#!/usr/bin/env bash
################################################################################
#
#   simple entrypoint script for hyperledger fabric
#
################################################################################

echo "arguments are $@"
echo "first arg is $1"

if [ ! -z "${WAIT_SIGNAL_FILES}" ] ; then
    wait-for-files.sh ${WAIT_SIGNAL_FILES}
fi

/tmp/register_users.sh $1
