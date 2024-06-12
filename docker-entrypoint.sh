#!/bin/bash

set -x
set -e

./relay-server.out \
  --listenPort="$LISTEN_PORT" \
  --externalEndpoint="$EXTERNAL_ENDPOINT"
