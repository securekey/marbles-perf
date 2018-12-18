#!/usr/bin/env bash

# SERVERS[CN]=HOSTNAMES
declare -A SERVERS=(
  [wild_vme_sk_dev]="*.vme.sk.dev"
)

# CLIENTS[CN]=HOSTNAMES
declare -A CLIENTS=(
  [fabric-dev]="fabric-client-dev"
)
