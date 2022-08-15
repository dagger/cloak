## Anatomy of a Dagger project

### Overview

All Dagger operations take place inside a project. A project is made of two parts:

* A *project file*
* A *context directory*. 

There are two ways to interact with a Dagger project:

* Invoking *workflows* from the CLI (`dagger do`) or Play With Dagger (web UI)
* Querying the *API* from client code.

### Workflows

Workflows are complete automations which can be invoked by the end user of a Dagger project.

Workflows are designed for convenience to a human operator, and are typically invoked from
the `dagger` CLI or the Play With Dagger web interface. Although workflows can be wrapped
by a third-party script, it's not recommended: better to query the Dagger API directly.


### API types

The Dagger API is a graphql-compatible API which can be used to inspect the capabilities of a
project, and compose those capabilities into powerful pipelines.

A project can add capabilities to the API with *API types*.


### How the Dagger CLI loads a project

Same logic as 'docker build':

* Default configuration file: `dagger.yaml`
* Default context directory: parent directory of the configuration file

### How Play With Dagger loads a project

In Play With Dagger, the loading logic is different since there is no local filesystem:

* The configuration file is edited directly in the browser, and stored in your Dagger Cloud account
* The context directory is downloaded from a git remote configured in the browser

### Project File Format

In addition to basic project metadata, the project file covers two major areas:
*extensions* and *SDK configuration*.

#### Extensions

Configuration key: `extensions`

Extensions can be installed into a project to add new capabilities. Extensions may add:

* workflows
* API types
* custom SDKs

To install a new extension:

1. Write a Dagger API query that returns the extension source
2. Append the query to the `extensions` field

Example:

```
extensions:
	- universe { yarn { extension } }
	- universe { netlify { extension } }
	- universe { go { extension(name: "go") } }
	- universe { go(version: "1.9") { extension(name: "go-1.9") } }
	- |
	  git {
	    pull(remote: "https://github.com/dagger/cloak", ref: "main") {
              subdir(path: "examples/todoapp") {
	        dagger {
		  extension(name: "todoapp-dev")
		}
	      }
	    }
          }
```

#### SDK configuration

Configuration key: `sdk`.

Most Dagger projects include some Dagger-specific code: custom API types, custom workflows, or both.
To build and load this code, a SDK (Software Development Kit) is required.

To use a SDK in your project:

* Choose the programming language you want to use
* Make sure the corresponding SDK is available. You may need to install an extension.
* Add your code to the project context directory
* Add an entry to the `sdk` key in your project file
* If needed, pass additional configuration to the SDK


In this example, the project has custom code in `.dagger/code`, with
Go and Typescript code sharing the same directory.

Note that the Go and Typescript SDKs require installing the Go and Typescript
extensions, respectively.

```
sdk:
  go:
    source: .dagger/code
    flags: -v
  typescript:
    source: .dagger/code

extensions:
  - universe { go { extension } }
  - universe { typescript { extension } }
```

#### Project file example

Here is an example of a complete project file:

```
name: todoapp
description: Automation workflows for the Dagger todoapp example
extensions:
	- universe { yarn { extension } }
	- universe { netlify { extension } }
	- universe { go { extension(name: "go") } }
	- universe { go(version: "1.9") { extension(name: "go-1.9") } }
	- |
	  git {
	    pull(remote: "https://github.com/dagger/cloak", ref: "main") {
              subdir(path: "examples/todoapp") {
	        dagger {
		  extension(name: "todoapp-dev")
		}
	      }
	    }
          }


sdk:
  go-1.9:
    source: .dagger/code
    flags: -v
  typescript:
    source: .dagger/code
```

Note the relationship between extension name and sdk name: if an extension implements an SDK,
that SDK will be available at the same name as the extension.
