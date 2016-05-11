# Cloud Foundry Service Broker for Redis Labs

## Installing the service

Download an archive from the [releases page](https://github.com/RedisLabs/cf-redislabs-broker/releases). Extract the contents.

## Running the service

Start the service pointing it to a config file:

```
redislabs-service-broker -c /path/to/config.yml
```

You can find a template for the config file in an `examples` [folder](https://github.com/RedisLabs/cf-redislabs-broker/tree/master/examples/config.yml). This template is distributed with every release as `config.yml.template`. Replace the values enclosed in `<>` with the actual parameter values. The properties not enclosed in `<>` are defaults that we find reasonable - you can alter them too.

## Using the service

Consult the [CF docs](http://docs.cloudfoundry.org/services/managing-service-brokers.html) to know more about managing brokers.

Some notes specific for the Redis Labs broker:

* A database name prefix is required for instance creation. Use the `-c` option to set one:
```
cf create-service ... -c '{"name":"mydatabase"}'
```
The instance ID is appended to this prefix in order to avoid name collisions. The name is then assigned to the DB in the RLEC API request.
* Any parameters described in the RLEC API docs can be specified via the `-c` option both on instance creation and instance update.
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
git clone https://github.com/RedisLabs/cf-redislabs-broker.git
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
./bin/test
```

### Adding a dependency

* Install [godep](https://github.com/tools/godep)
* Install the dependency (eg via `go get`) and ensure everything works fine
* `godep save ./...`
* Check that the output of `git diff vendor/ Godeps/` looks reasonable
* Commit `vendor/` and `Godeps/`
