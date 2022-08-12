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
// Optional project name
name: string

// Optional project description
description?: string


// A cloak pipeline encoded as a grapqhl query
#pipeline: string

// Useful API queries available to all clients
queries: {
    // Action pipelines invoked by 'dagger do'
    do: [string]: #pipeline

    // Artifact pipelines invoked by 'dagger pull'
    pull: [string]: #pipeline

    // Service pipelines invoked by 'dagger run'
    run: [string]: #pipeline
}

// Implement custom types to extend the API
types: [
    ...{
        // Source directory for the type implementation
        source: #pipeline

        // SDK to use to build the type implementation
        // FIXME: custom SDK
        sdk: "go" | "typescript" | "bash" | "python"
    }
]

// Load extensions
extensions: [
    ...{
        // Optionally override extension name
        name?
        // Pipeline to retrieve the extension source
        source: #pipeline
    }
]

```

### Project file examples

```yaml
queries:
    do:
        deploy: |
            {
                host { getenv(key: "NETLIFY_TOKEN") { save("token") } }
                yarn { build { save("bld") } }
                netlify { deploy(contents: "bld") }
            }

types:
    -
        source: { core { subdir(path: "examples/todoapp/go", sdk: "go" ) } }

extensions:
    - source: universe { yarn }
    - source: universe { netlify }
```