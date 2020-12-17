#!/bin/bash

set -euo pipefail

#LAV LATEST_ARTIFACT_VERSION
LAV="v0.2.8"

curl -s -L --retry 5 https://github.com/semaphoreci/artifact/releases/download/$LAV/artifact_Linux_x86_64.tar.gz -o Linux.tar.gz

curl -s -L --retry 5 https://github.com/semaphoreci/artifact/releases/download/$LAV/artifact_Darwin_x86_64.tar.gz -o Darwin.tar.gz

git clone git@github.com:semaphoreci/toolbox.git

tar -zxf Linux.tar.gz

mv artifact toolbox/

FILE_LIST=""
for FILE in toolbox/*; do
  FILE_LIST="$FILE_LIST $FILE"
done

tar -cf toolbox_Linux.tar $(echo $FILE_LIST)

echo "toolbox Linux content: "
tar --list --verbose --file=toolbox_Linux.tar

rm toolbox/artifact

tar -zxf Darwin.tar.gz

mv artifact toolbox/

FILE_LIST=""
for FILE in toolbox/*; do
  FILE_LIST="$FILE_LIST $FILE"
done

tar -cf toolbox_Darwin.tar $(echo $FILE_LIST)

echo "toolbox Darwin content: "
tar --list --verbose --file=toolbox_Darwin.tar

latest=$(git tag | sort --version-sort | tail -n 1)

release_id=$(curl --silent "https://api.github.com/repos/semaphoreci/toolbox/releases/tags/$latest" | grep -m1 'id' | awk '{print $2}' | tr -d ',' )

curl \
    -X POST \
    -H "Authorization: token $GITHUB_TOKEN" \
    -H "Accept: application/vnd.github.v3+json" \
    -H "Content-Type: $(file -b --mime-type toolbox_Linux.tar)" \
    --data-binary @toolbox_Linux.tar \
    "https://uploads.github.com/repos/semaphoreci/toolbox/releases/$release_id/assets?name=toolbox_Linux.tar"

echo "toolbox_Linux.tar uploaded"

curl \
    -X POST \
    -H "Authorization: token $GITHUB_TOKEN" \
    -H "Accept: application/vnd.github.v3+json" \
    -H "Content-Type: $(file -b --mime-type toolbox_Darwin.tar)" \
    --data-binary @toolbox_Linux.tar \
    "https://uploads.github.com/repos/semaphoreci/toolbox/releases/$release_id/assets?name=toolbox_Darwin.tar"

echo "toolbox_Darwin.tar uploaded"
