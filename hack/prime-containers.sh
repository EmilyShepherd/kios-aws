#!/bin/bash
#

podman --root datapart/data/oci pull --platform=linux/amd64 \
  registry.k8s.io/pause:3.6 \
  $(grep image: datapart/meta/etc/kubernetes/manifests/*.yaml | grep -o '[^ ]\+$') \
  "$@"

cd datapart/data/oci

find -user $USER -type f | xargs chmod u+r
find -user $USER -type d | xargs chmod u+rx

rm -rf overlay-containers libpod overlay-layers/*.gz defaultNetworkBackend *.lock */*.lock
