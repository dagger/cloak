---
author: Solomon Hykes <solomon@dagger.io>
status: proposal
implemented: no
---

# Configuring a cloak project

## Overview

All cloak operations take place inside a *project*. A project is a directory that includes a valid *project configuration file*. The configuration file contains all the information necessary for cloak to build, load and run the project's automation API.

## The cloak project file

Any directory that contains a file named `cloak.yaml` is considered a cloak project. A project may contain other projects, but when loading a project, its parent projects are ignored.

### Project file schema

The project file must 1) be encoded in yaml, and 2) match the following schema (expressed in [Cue](https://cuelang.org) for readability.

```cue
// Schema version
schemaVersion?: "0.0.1" | *"latest"

// Optional project name
name: string

// Optional project description
description?: string

// Add capabilities to your API with code

#codeDir: {
    source: #Source
    sdk: "go" | "bash" | ...
}


code?: #codeDir | [...#codeDir]

// Recursively load other projects, and include their API in ours, with a namespace
extensions: [
    ... {
        // Override the default extension name
        name?: string
        

        // The source directory for the extension.
        // Extensions are basically imported projects.
        // They are loaded recursively, and their API stitched in, with
        // each extension's actions and types in their own namespace.
        source: #Source

        // FIXME: extension settings?
    }
]

// All the different ways to configure a source directory.
#Source: {
        {
            // Load the extension source from a subdirectory of the project
            subdir: string
        } | {
            // Load the extension source from a directory on the local host
            // Note: this is a privileged operation and must require trusting the project author
            localdir: string
        } | {
            // Download the source from a git repository
            git: {
                remote: string
                ref?: string
                dir: string
            }
        } | {
            // Download the source from an OCI registry
            oci: {
                ref: string
                digest?: string
            }
        } | {
            // Download the source as a tar archive over http
            http: string
        } | {
            // execute a cloak pipeline that outputs the source directory
            // pipeline is encoded as a graphql query and executed by cloak
            pipeline: string
        }
```

### Project file examples

```yaml
actions:
    deploy:
        source:
            subdir: .dagger/actions/deploy
        builder: "core { dockerbuild { source: $source }}"
types:
    -
        source:
            subdir: .dagger/types
    -
        source:
            git:
                remote: https://github.com/MYORG/PLATFORM
                ref: stable
                subdir: dagger/types/common
extensions:
    yarn:
        source:
            git:
                remote: https://github.com/dagger/cloak
                ref: main
                dir: examples/yarn


extensions:
    -
```
