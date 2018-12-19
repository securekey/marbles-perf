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

SHELL := /bin/bash

ifeq ($(GOPATH),)
    GOPATH = ${HOME}/go
endif

# environment variables for the adapter builds
BUILD_HASHTAG ?= $(shell git describe --tags)
BUILD_DATE ?= $(shell date +%Y%m%d.%H%M%S)
BUILD_COMMIT ?= $(shell git rev-parse HEAD)
BUILD_BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD)
CHANGES_PENDING ?= $(shell git status -s | grep -c ".*")
ADAPTER_BASEOS_IMAGE ?= $(or $(docker_baseos_image_name),adapter-baseos)

export WORKSPACE_TEST_OUTPUT_DIR?=/tmp/test-results
export MARBLES_PERF_VERSION ?= $(shell git rev-parse --short=7 HEAD)
#IMAGE_NAME_PREFIX ?= repo.onetap.ca:8444/vme

ifeq ($(MARBLES_PERF_VERSION),)
  DOCKER_IMAGES_TAG=latest
else
  DOCKER_IMAGES_TAG=snapshot-$(MARBLES_PERF_VERSION)
endif

# Builds for docker images -----------------------------------------------------
# Maintained by Service Engineering team

# function used to build docker images
define docker_build_image
    docker build -t $(1):$(DOCKER_IMAGES_TAG) -f $(2) .
endef

docker: populate docker-marbles-perf

#docker-push-all:

docker-marbles-perf:
	@echo "Building marbles-perf service docker image"
	$(call docker_build_image,marbles-perf,images/marbles-perf/Dockerfile)

docker-clean-dangling:
	@echo "Removing dangling images"
	-docker rmi -f $$(docker images -f "dangling=true" -q)

services-up: populate start/marbles-perf

restart-marbles-perf: stop/marbles-perf start/marbles-perf

restart-marbles-perf: stop/marbles-perf start/marbles-perf

start/marbles-perf:
	@cd deployment/compose && MARBLES_PERF_TAG=$(DOCKER_IMAGES_TAG) docker-compose -p marbles-perf -f compose-marbles-perf.yml up -d

stop/marbles-perf:
	@cd deployment/compose && MARBLES_PERF_TAG=$(DOCKER_IMAGES_TAG) docker-compose -p marbles-perf -f compose-marbles-perf.yml stop

.PHONY: populate-vendors populate

populate-vendors:
	@./populate_vendor.sh

populate: populate-vendors


fabric-up:
	@cd deployment/fabric && . compose.env && COMPOSE_DIR=$(pwd) ./network_setup.sh up

fabric-down:
	@cd deployment/fabric && . compose.env && COMPOSE_DIR=$(pwd) ./network_setup.sh down

