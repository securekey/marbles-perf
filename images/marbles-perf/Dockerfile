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
