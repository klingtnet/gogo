# gogo

`gogo` is a wrapper around the _go tool_ which manages a local workspace for your project.

## Usage

Change into the root directory of your go project and boostrap the local workspace: `gogo bootstrap <import-path>` where `<import-path>` is something like 'github.com/klingtnet/gogo'.
Now you can run `gogo <go-command> [argument]...` from any subdirectory of your project, e.g. `gogo build ./...`.
`gogo` will take care of setting the `GOPATH` to the project's workspace before running any go command inside the appropriate `$GOPATH/src` directory.

## Differences to Existing Tools

- TODO: `gb` and `wgo`
