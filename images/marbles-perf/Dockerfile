#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
#

#
#  BUILD CONTAINER
#

# Container used to build the binaries
FROM   golang:alpine as builder

# Source code directories
ENV    MARBLES_PERF_SRC="${GOPATH}/src/github.com/securekey/marbles-perf"

# Copy source code
RUN    set -xe; mkdir -p ${MARBLES_PERF_SRC}
COPY   . ${MARBLES_PERF_SRC}/

#
#  Compiling
#

# Building adapter
ARG    go_extra_args=
ARG    build_hashtag=
ARG    build_date=
ARG    build_commit=
ARG    build_branch=
ARG    changes_pending=

RUN    set -x; \
       cd ${MARBLES_PERF_SRC}/service && go build -o /marbles-perf

#
#  FINAL CONTAINER
#

FROM   alpine:3.8

# default user id of this docker image
ARG    default_image_uid=30201

# Jenkins build ID
ARG    build_id=UNSET

# app Home
ARG    app_home=/opt/securekey/marbles-perf

# LABELS
LABEL  vendor="SecureKey Technologies Inc." \
       com.securekey.build_id=${build_id}

# SET SUPPORT_TOOL DEFAULTS
ENV    APP_HOME=${app_home} \
       APP_CFG=${app_home}/config.yaml

# ENVIRONMENT VARIABLES TO SET TIMEZONE
ENV    TZ=America/Toronto


#
# ADDING FILES INTO THE CONTAINER
#

# Copying compiled binaries and entrypoint
COPY   --from=builder /marbles-perf /usr/local/bin
COPY   images/marbles-perf/entrypoint.sh /usr/local/bin/entrypoint.sh

# Installing required packages
RUN    set -xe; \

       # Creating directory structure
       mkdir -p ${app_home} && \

       # Make the binaries executable and not writable
       chmod 555 /usr/local/bin/*

# Copying files required for the support tool
COPY   images/marbles-perf/homedir ${app_home}/

# Fixing permissions - Allow access to only SK_DEFAULT_USER and root
RUN    chown -R ${default_image_uid}:root ${app_home} && \
       chmod -R 770 ${app_home}

# EXPOSE PORTS
EXPOSE 8080

# DEFAULT USER IS SET TO ${default_image_uid}
USER   ${default_image_uid}:0

# ENTRY
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
