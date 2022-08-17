# Project file

`dagger.yaml`

## Concepts

- A project file defines workflows
- A workflow makes a query to the API
- A query can have optional parameters
- A query can be bound to optional host resources (eg. local directory, env vars, secrets)
- A project file loads extensions
- An extension can come from 3 sources
  - core: builtins implemented by dagger
  - universe: public packages from the catalog
  - local: extensions local to my project (under `.dagger/extensions/<name>`)

## Other principles

- An extension targets a language SDK and can call other extensions
- If my project's workflow needs to call other extensions and define a DAG, I should write my own extension in my language of choice
- A workflow is "just" an entrypoint to one or several extensions, it can only specify a query but cannot define a DAG
- A workflow has access to privileged host resources (an extension does not, it can only receive them as arguments)

## Example of a project file

Example of a project file for the dagger docs website:

```yaml
name: "dagger-docs"

# Extensions to load, shared by all workflows (whether they need it or not)
# optionally each workflow could specify extra dependencies
extensions:
  # Specifies a local extension (under the directory ".dagger/extensions/")
  local:
    - myGoExtension
  # Specifies public extensions from the Universe catalog
  universe:
    - yarn

workflows:
  # run: dagger do test
  - name: test
    # Source directory to pass to the workflow
    source: ./tests
    # API query describing the workflow behavior (could be serialized with `yarn: script: tests` to avoid typing gql queries)
    query: |
      {
        universe {
          yarn(src: $source, script: "tests")
        }
      }
    # Pass local environment variables to the workflow
    env:
      - DEBUG

  # The dev.sh script simply runs "yarn install && yarn start"
  # the only advantage here is to run it from a container instead of on the host
  # run: dagger do dev
  - name: dev
    # source could be omitted if defaults to "."?
    source: "."
    query:
      yarn:
        - src: $source
        - script: start

  # run: dagger do deploy
  - name: deploy
    source: "."
    # Shortcut for `{ core { deployDocs(src: $source) } }`
    query: |
      {
        local {
          myGoExtension {
            deployDocs(src: $source, token: $netlifyToken)
          }
        }
      }
    parameters:
      #FIXME: how to secret?
      - netlifyToken: $NETLIFY_TOKEN
```

The code of the local extension `myGoExtension` is close (if not identical) to [`./examples/todoapp/go/`](https://github.com/dagger/cloak/tree/f28041072e21dbdaa533be2cd2e2987a84aa7d4f/examples/todoapp/go)
