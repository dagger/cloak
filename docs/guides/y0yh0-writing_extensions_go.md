---
slug: /y0yh0/writing_extensions_go
displayed_sidebar: "0.3"
---

# Writing a new project with Go

Say we are creating a new project called `foo`. It will have

1. A single extension, written in Go, the extends the schema with an action called `bar`.
1. A script, also written in Go, that can call the extension (and any other project dependencies)

## Setup the project configuration

1. Enter a go module directory for your project (`go mod init <module name>` if one doesn't exist)
1. Configure go to use the private cloak git repo (this will go away once the repo is made public)

   - If not already set, this command will update your `~/.gitconfig` with a rule that tells git to use ssh instead of https for the cloak repo:

     - `git config --global --add url.ssh://git@github.com/dagger/cloak.insteadOf https://github.com/dagger/cloak`

   - Then run the following commands to setup the rest of the required dependencies

     ```console
     export GOPRIVATE=github.com/dagger/cloak
     go get github.com/dagger/cloak@main
     # This is needed to fix a transitive dependency issue (`sirupsen` vs. `Sirupsen`...)
     go mod edit -replace=github.com/docker/docker=github.com/docker/docker@v20.10.3-0.20220414164044-61404de7df1a+incompatible
     ```

1. In order to pull cloak dependencies and build the extension in this example, cloak will need pull the private repo from a container running in the engine.

   - Setting up an ssh-agent with credentials that can pull the `dagger/cloak` will cover all these cases and is recommended for now.
     - Github has [documentation on setting this up for various platforms](https://docs.github.com/en/authentication/connecting-to-github-with-ssh/generating-a-new-ssh-key-and-adding-it-to-the-ssh-agent#adding-your-ssh-key-to-the-ssh-agent).
     - Be sure that the `SSH_AUTH_SOCK` variable is set in your current terminal (running `eval "$(ssh-agent -s)"` will typically take care of that)
     - Without this, you may get error messages containing `no ssh handler for id default`
   - Alternatively, if you don't have to pull any cloak dependencies (e.g. your `dependencies` key in `cloak.yaml` is empty), you can avoid the need to setup ssh-agent by vendoring your dependencies (`go mod vendor`)

1. Create a new file called `cloak.yaml`

   - This is where you declare your project, and other project that it depends on. All extensions declared in this file will be built, loaded, and available to be called when the project is loaded into cloak.
   - Create the file in the following format:

   ```yaml
   name: foo
   scripts:
     - path: script
       sdk: go
   dependencies:
     - git:
         remote: git@github.com:dagger/cloak.git
         ref: main
         path: examples/yarn/cloak.yaml
     - git:
         remote: git@github.com:dagger/cloak.git
         ref: main
         path: examples/netlify/go/cloak.yaml
   ```

   - The dependencies are optional and just examples, feel free to change as needed.
   - `core` does not need to be explicitly declared as a dependency; it is implicitly included. If your only dependency is `core`, then you can just skip the `dependencies:` key entirely.

## Create your script

### Generate initial script code

1. From the directory containing `cloak.yaml` (or any subdirectory thereof), run `cloak generate`
1. You should now see:

   - A `script/main.go` file

1. NOTE: if you re-run `cloak generate` in the future, it will refuse to overwrite your existing `main.go` file if it exists (in order to not clobber your work). For now, if you update your schema you won't get new autogenerated skeletons for it unless you temporarily rename your `main.go` file. We intend to make this less tedious with future SDK updates.

### Implement the script

1. Edit `script/main.go`, replacing the `panic("implement me")` with your actual implementation.
1. When you need to call a dependency declared in `cloak.yaml`, you will currently have to use raw graphql queries. Examples of this can be found in [the alpine extension here](https://github.com/dagger/cloak/blob/main/examples/alpine/main.go).
1. Also feel free to import any other third-party libraries as needed the same way you would with any other go project.

### Invoke your script

1. If this is the first time running it, you may need a `go mod tidy` to apply the previous go mod commands
1. The simplest way to invoke is to execute `go run script/main.go`

## Create your extension

Update your `cloak.yaml` to include a new `extensions` key:

```yaml
name: foo
scripts:
  - path: script
    sdk: go
extensions:
  - path: ext
    sdk: go
dependencies:
  - git:
      remote: git@github.com:dagger/cloak.git
      ref: main
      path: examples/yarn/cloak.yaml
  - git:
      remote: git@github.com:dagger/cloak.git
      ref: main
      path: examples/netlify/go/cloak.yaml
```

### Create schema file

- Create a new file `ext/schema.graphql`, which will define the new APIs implemented by your extension and vended by your project.

  - Example contents for a single `bar` action:

    ```graphql
    extend type Query {
      foo: Foo!
    }

    type Foo {
      bar(in: String!): String!
    }
    ```

  - Also see other examples:
    - [alpine](https://github.com/dagger/cloak/blob/main/examples/alpine/schema.graphql)
    - [netlify](https://github.com/dagger/cloak/blob/main/examples/netlify/go/schema.graphql)
  - NOTE: this step may become optional in the future if code-first schemas are supported

### Generate initial extension code

1. From any project directory (that is, the directory containing `cloak.yaml` or any subdirectory thereof), run `cloak generate`

1. You should now see:
   - A `ext/main.go` file
   - Some autogenerated boilerplate: structures for needed types in `models.go` and some runtime code that makes your extension invokable in the containerized cloak engine in `generated.go`
1. NOTE: this has the same behavior as scripts in that `cloak generate` will refuse to overwrite any existing `main.go` file (mentioned above)

### Implement the extension

1. Edit `ext/main.go`, replacing occurences of `panic("implement me")` with the implementation of your extension's actions.
1. When you need to call a dependency declared in `cloak.yaml`, you will currently have to use raw graphql queries. Examples of this can be found in [the alpine extension here](https://github.com/dagger/cloak/blob/main/examples/alpine/main.go).
1. Also feel free to import any other third-party dependencies as needed the same way you would with any other go project. They should all be installed and available when executing in the cloak engine.
1. Some examples:
   - [alpine](https://github.com/dagger/cloak/blob/main/examples/alpine/main.go)
   - [netlify](https://github.com/dagger/cloak/blob/main/examples/netlify/go/main.go)

### Invoke your extension

1. One simple way to verify your extension builds and can be invoked is via the graphql playground.
   - Just run `cloak dev` from any directory in your project and navigate to `localhost:8080` in your browser (may need [an SSH tunnel](https://www.ssh.com/academy/ssh/tunneling-example) if on a remote host)
     - you can use the `--port` flag to override the port if needed
   - Click the "Docs" tab on the right to see the schemas available, including your extension and any dependencies.
   - You can submit queries by writing them on the left-side pane and clicking the play button in the middle
1. You can also use the cloak CLI, e.g.

   ```console
   cloak do <<'EOF'
   {
     foo {
       bar(in: "in")
     }
   }
   EOF
   ```

1. Finally, you should now be able to invoke your extension from your script too.
