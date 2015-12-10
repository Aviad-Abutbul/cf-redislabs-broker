# Cloud Foundry Service Broker for Redis Labs

## Installing the service

Download an archive from the [releases page](https://github.com/Altoros/cf-redislabs-broker/releases). Extract the binary.

## Running the service

Start the service pointing it to a config file:

```
redislabs-service-broker -c /path/to/config.yml
```

You can find a template for the config file in an `examples` [folder](https://github.com/Altoros/cf-redislabs-broker/tree/master/examples/config.yml). Replace the values enclosed in `<>` with the actual parameter values. The properties not enclosed in `<>` are defaults that we find reasonable - you can alter them too.

## Using the service

Consult the [CF docs](http://docs.cloudfoundry.org/services/managing-service-brokers.html) to know more about managing brokers.

Some notes specific for the Redis Labs broker:

* Additional parameters (the `-c` option) are not supported on service instance creation. See the information about updating service instances in the next item.
* The following parameters can be updated on an instance update (refer to the [RLEC docs](https://redislabs.com/redis-enterprise-documentation/overview) for details):
  - `memory_size` (integer, expressed in bytes)
  - `data_persistence` (string, either "disabled", "aof", or "snapshot")
  - `snapshot_policy` (array of objects each with 2 properties - `writes` and `secs`, both integers)
* You can switch between plans on a service update. Moreover, you can both switch between plans and update parameters described above at the same time.
* The broker works in a synchronous way - all the time you just need to wait until the command has finished. Note that there is a 15 seconds timeout awaiting for a database creation - if it is over the request would fail.

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
