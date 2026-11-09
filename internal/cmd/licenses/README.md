# licenses

The `licenses` command checks the licenses of the dependencies compiled into the provider and generates the third-party notices.

It is based on Google's [go-licenses](https://github.com/google/go-licenses). The `check` command runs `go-licenses check`, and `generate` runs `go-licenses report` and `go-licenses save` for each platform and merges the results. The `go-licenses` version is pinned as a tool dependency in `go.mod`.

## Make targets

Use these `make` targets from the repository root:

- `make license-check` fails if a dependency uses a license that isn't allowed.
- `make license-gen` regenerates `THIRD_PARTY_NOTICES.md` and saves the license texts to `LICENSES/`. Run it after you change the dependencies, and commit the updated `THIRD_PARTY_NOTICES.md`.
- `make license-gen-check` runs `make license-gen` and fails if that changes any tracked file. It confirms that the committed `THIRD_PARTY_NOTICES.md` matches the current dependencies.

CI runs `make license-check` and `make license-gen-check` when `go.mod`, `go.sum`, the notices, or the license tooling change.

## Configuration

`config.yml` holds the allowed licenses (`allowed-licenses`) and the packages to leave out of the check (`ignored-packages`). Pass another file with `--config`.

## Commands

Run the command directly to use other options:

```shell
go run ./internal/cmd/licenses check
go run ./internal/cmd/licenses generate
```

Both commands accept `--platform` (repeatable) to limit the `GOOS/GOARCH` platforms. The default is every platform built in `.goreleaser.yml`. `generate` also accepts `--save-licenses-dir` to save the license texts, and `--include-mpl-source` to keep the source of MPL-2.0 dependencies in that directory. Each flag can also be set through an environment variable, for example `PLATFORM`.
