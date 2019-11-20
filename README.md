# drone-bintray

[![Build Status](http://cloud.drone.io/api/badges/drone-plugins/drone-bintray/status.svg)](http://cloud.drone.io/drone-plugins/drone-bintray)
[![Gitter chat](https://badges.gitter.im/drone/drone.png)](https://gitter.im/drone/drone)
[![Join the discussion at https://discourse.drone.io](https://img.shields.io/badge/discourse-forum-orange.svg)](https://discourse.drone.io)
[![Drone questions at https://stackoverflow.com](https://img.shields.io/badge/drone-stackoverflow-orange.svg)](https://stackoverflow.com/questions/tagged/drone.io)
[![](https://images.microbadger.com/badges/image/plugins/bintray.svg)](https://microbadger.com/images/plugins/bintray "Get your own image badge on microbadger.com")
[![Go Doc](https://godoc.org/github.com/drone-plugins/drone-bintray?status.svg)](http://godoc.org/github.com/drone-plugins/drone-bintray)
[![Go Report](https://goreportcard.com/badge/github.com/drone-plugins/drone-bintray)](https://goreportcard.com/report/github.com/drone-plugins/drone-bintray)

> Warning: This plugin has not been migrated to Drone >= 0.5 yet, you are not able to use it with a current Drone version until somebody volunteers to update the plugin structure to the new format.

Drone plugin to publish files and artifacts to Bintray. For the usage information and a listing of the available options please take a look at [the docs](http://plugins.drone.io/drone-plugins/drone-bintray/).

## Build

Build the binary with the following command:

```console
export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=0
export GO111MODULE=on

go build -v -a -tags netgo -o release/linux/amd64/drone-bintray
```

## Docker

Build the Docker image with the following command:

```console
docker build \
  --label org.label-schema.build-date=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
  --label org.label-schema.vcs-ref=$(git rev-parse --short HEAD) \
  --file docker/Dockerfile.linux.amd64 --tag plugins/bintray .
```

## Usage

```console
docker run --rm \
       -e PLUGIN_BINTRAY_USERNAME=$BINTRAY_USER \
       -e PLUGIN_BINTRAY_API_KEY=$BINTRAY_KEY \
       -e PLUGIN_BINTRAY_CFG=./package_config_new.yaml \
       -e PLUGIN_BINTRAY_GPG_PASSPHRASE=$BINTRAY_ADMIN_GPG_PASSPHRASE \
       -v $(pwd):$(pwd) \
       -w $(pwd) \
       plugins/bintray

OR

docker run --rm \
       --env-file .env \
       -v $(pwd):$(pwd) \
       -w $(pwd) \
       plugins/bintray
```
