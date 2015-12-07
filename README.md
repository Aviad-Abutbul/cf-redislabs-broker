# Cloud Foundry Service Broker for Redis Labs

## Installing the service

TODO create a release

## Using the service

Start the service pointing it to a config file:

```
redislabs-service-broker -c /path/to/config.yml
```

You can find a template for the config file in an `examples` [folder](https://github.com/Altoros/cf-redislabs-broker/tree/master/examples/config.yml). Replace the values enclosed in `<>` with the actual parameter values.

### Logs

The program logs DEBUG-level info to `stdout` and errors to `stderr`.

### Internal state

The broker stores its state in a JSON file located in a `$HOME/.redislabs-broker` folder. NOTE: Do not change the contents of this folder manually.

The persistence is implemented as a pluggable backend. Therefore, an option of storing the state in a SQL/NoSQL database may appear soon in the future.

## Development

### Configuring the environment

* Get the code:
```
git clone https://github.com/Altoros/cf-redislabs-broker.git
cd cf-redislabs-broker
```
* Install Go 1.5
* [Ensure your GOPATH is set correctly](https://golang.org/cmd/go/#hdr-GOPATH_environment_variable)
* In managing dependencies, we rely on Go 1.5 Vendor Experiment. Therefore, set up a `GO15VENDOREXPERIMENT` variable to equal `1`. You can use `./bin/go` to have it set up for you.

### Building the binary

```
./bin/build
```

After that you'll find the resulting binary in `out/redislabs-service-broker`.

### Running unit tests

```
go test ./redislabs/...
```

### Adding a dependency

* Install [godep](https://github.com/tools/godep)
* Install the dependency (eg via `go get`) and ensure everything works fine
* `godep save ./...`
* Check that the output of `git diff vendor/ Godeps/` looks reasonable
* Commit `vendor/` and `Godeps/`
