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

// Optional custom actions
// Note: use this for statically defined actions.
// Additional actions may be loaded dynamically by hooks.
actions?: {
    [name=string]: {
        // Action source directory
        source: #Source

        // Pipeline to build action from source.
        // encoded as a graphql query and executed by cloak.
        // Default is standard docker build.
        // Action source is available to the query as the gql variable `$source`
        builder?: string | *"core { dockerbuild { source: $source } }"

        // FIXME: sub-actions?
    }
}

// Optional custom types
types?: [
    ...{
        {
            // Load graphql types from a source directory
            source: #Source
            // Glob pattern specifying which files to match
            glob: string | *"*.gql"
        } | {
            // Inline graphql schema
            schema: string
        }
    }
]

// FIXME: services
// FIXME: artifacts

// Extensions to load into the project
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
