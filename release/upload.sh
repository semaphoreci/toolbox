#!/bin/bash
set -euo pipefail

#
# Upload tarballs to Github.
#
# How to use:
#
#   1. Create the tarballs by running release/create.sh -a, this will create:
#      - /tmp/Linux/linux.tar
#      - /tmp/Darwin/darwin.tar
#      - /tmp/self-hosted-Linux/linux.tar
#      - /tmp/self-hosted-Darwin/darwin.tar
#
#   2. Upload the tarballs to Github by running release/upload.sh.
#

echo "Uploading to GitHub"

curl \
  -X POST \
  -H "Authorization: token $GITHUB_TOKEN" \
  -H "Accept: application/vnd.github.v3+json" \
  https://api.github.com/repos/semaphoreci/toolbox/releases \
  -d '{"tag_name":"'$SEMAPHORE_GIT_TAG_NAME'"}'

release_id=$(curl --silent https://api.github.com/repos/semaphoreci/toolbox/releases/tags/$SEMAPHORE_GIT_TAG_NAME | grep -m1 'id' | awk '{print $2}' | tr -d ',' )

curl \
    -X POST \
    -H "Authorization: token $GITHUB_TOKEN" \
    -H "Accept: application/vnd.github.v3+json" \
    -H "Content-Type: $(file -b --mime-type /tmp/Linux/linux.tar)" \
    --data-binary @/tmp/Linux/linux.tar \
    "https://uploads.github.com/repos/semaphoreci/toolbox/releases/$release_id/assets?name=linux.tar"

echo "linux.tar uploaded"

curl \
    -X POST \
    -H "Authorization: token $GITHUB_TOKEN" \
    -H "Accept: application/vnd.github.v3+json" \
    -H "Content-Type: $(file -b --mime-type /tmp/self-hosted-Linux/linux.tar)" \
    --data-binary @/tmp/self-hosted-Linux/linux.tar \
    "https://uploads.github.com/repos/semaphoreci/toolbox/releases/$release_id/assets?name=self-hosted-linux.tar"

echo "self-hosted-linux.tar uploaded"

curl \
    -X POST \
    -H "Authorization: token $GITHUB_TOKEN" \
    -H "Accept: application/vnd.github.v3+json" \
    -H "Content-Type: $(file -b --mime-type /tmp/Darwin/darwin.tar)" \
    --data-binary @/tmp/Darwin/darwin.tar \
    "https://uploads.github.com/repos/semaphoreci/toolbox/releases/$release_id/assets?name=darwin.tar"

echo "darwin.tar uploaded"

curl \
    -X POST \
    -H "Authorization: token $GITHUB_TOKEN" \
    -H "Accept: application/vnd.github.v3+json" \
    -H "Content-Type: $(file -b --mime-type /tmp/self-hosted-Darwin/darwin.tar)" \
    --data-binary @/tmp/self-hosted-Darwin/darwin.tar \
    "https://uploads.github.com/repos/semaphoreci/toolbox/releases/$release_id/assets?name=self-hosted-darwin.tar"

echo "self-hosted-darwin.tar uploaded"

curl \
    -X POST \
    -H "Authorization: token $GITHUB_TOKEN" \
    -H "Accept: application/vnd.github.v3+json" \
    -H "Content-Type: $(file -b --mime-type /tmp/self-hosted-Windows/windows.tar)" \
    --data-binary @/tmp/self-hosted-Windows/windows.tar \
    "https://uploads.github.com/repos/semaphoreci/toolbox/releases/$release_id/assets?name=self-hosted-windows.tar"

echo "self-hosted-windows.tar uploaded"

echo "Everything DONE"
echo "🍻"
