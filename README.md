# DV backend

## Linters

To format all the source texts of the project in accordance
with the [Go](https://go.dev/) standard, you can use the command:

```shell
go fmt ./...
```

The following linters are used in the project:

- [go vet](https://pkg.go.dev/cmd/vet)
- [errcheck](https://github.com/kisielk/errcheck)  
  command install linter: `go install github.com/kisielk/errcheck@latest`.
- [staticcheck](https://staticcheck.io/)  
  command install linter: `go install honnef.co/go/tools/cmd/staticcheck@latest`.
- [usestdlibvars](https://github.com/sashamelentyev/usestdlibvars)  
  command install linter: `go install github.com/sashamelentyev/usestdlibvars@latest`

All linters of the project can be run with the command:

```shell
make lint
```

## Testing

To unit-tests, you need to run the command in the root of the project:

```shell
make test
```

## Building

[Go](https://go.dev/) version 1.22.0 or higher is required for the build.

To build, you need to run the command in the root of the project:

```shell
make build
```

After completing the command, the binary file `github.com/dv-net/dv-merchant` will appear in the
`.bin` folder.

## Running

Start project for developer

```shell
cd .docker/dev
docker-compose up -d --no-deps --build github.com/dv-net/dv-merchant
```

Or using `make run`.

## SQL

To generate SQL wrappers you need to install `pgxgen`:

```shell
go install github.com/tkcrm/pgxgen/cmd/pgxgen@latest
```

## CLI commands

#### Start app server
```bash
github.com/dv-net/dv-merchant start
```

#### Show app version
```bash
github.com/dv-net/dv-merchant version
```
#### Run db migrate
```bash
github.com/dv-net/dv-merchant migrate up/down
```
#### Run db seed
```bash
github.com/dv-net/dv-merchant seed up/down
```
#### Run config validate, gen envs and flags for config
```bash
github.com/dv-net/dv-merchant config 
```
#### Run permission command
```bash
github.com/dv-net/dv-merchant permission 
```

#### Transactions management
```bash
github.com/dv-net/dv-merchant transactions 
```

#### Users management
```bash
github.com/dv-net/dv-merchant users 
```