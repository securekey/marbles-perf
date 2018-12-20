#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
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

services-down: stop/marbles-perf

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

