# redislabs-service-broker


## TODOs

- [ ] Configuration management including:
  - [ ] admin creds
  - [ ] plans list
  - [ ] nats config
  - [ ] node list
  - [ ] under the question: database connection for persistence (redis or psql)
- [ ] Broker API
  - [ ] services (list plans)
  - [ ] provision
  - [ ] deprovision
  - [ ] bind
  - [ ] ? unbind
- [ ] User management
- [ ] Logging

## Development

### Configuring the environment

* Get the code:
```
git clone https://github.com/Altoros/cf-redislabs-broker.git
cd cf-redislabs-broker
```
* Install Go 1.5
* [Ensure your GOPATH is set correctly](https://golang.org/cmd/go/#hdr-GOPATH_environment_variable)
* In managing dependencies, we rely on Go 1.5 Vendor Experiment. Therefore, set up a `GO15VENDOREXPERIMENT` variable to equal `1`. `/bin/env` script can do it for you.

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

* Install [glide](https://github.com/Masterminds/glide.git)
* `glide get github.com/<ORG>/<REPO>`
* `find vendor -name '.git' | xargs rm -r`
* `git add vendor glide.yaml` & commit

Note: use `glide` with at least `0.7.2` version.

### Starting the service

```
./out/redislabs-service-broker -c config.yml
```

You can find a sample of the config file in an `examples` [folder](https://github.com/Altoros/cf-redislabs-broker/tree/master/examples).
