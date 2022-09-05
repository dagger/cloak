#!/bin/bash

set -ex

# https://archive.ph/hzM6G

URL="cloudnativeday.ch"

# https://www.petekeen.net/archiving-websites-with-wget
wget \
    --mirror \
    --warc-file="$URL" \
    --warc-cdx \
    --page-requisites \
    --html-extension \
    --convert-links \
    --execute robots=off \
    --directory-prefix=. \
    --span-hosts \
    --domains="$URL",js.tito.io \
    --wait=1 \
    --random-wait \
    "$URL"
