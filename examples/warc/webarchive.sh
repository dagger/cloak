#!/bin/bash

set -ex

# https://archive.ph/hzM6G

URL="https://cloudnativeday.ch/an-electric-automation-engine"

# https://www.petekeen.net/archiving-websites-with-wget
wget \
    --mirror \
    --warc-file="cloudnative.ch" \
    --warc-cdx \
    --page-requisites \
    --html-extension \
    --convert-links \
    --execute robots=off \
    --directory-prefix=. \
    --span-hosts \
    --domains=cloudnative.ch,js.tito.io \
    --wait=1 \
    --random-wait \
    "$URL"
