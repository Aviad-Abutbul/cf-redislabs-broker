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

### How to build

```
git clone https://github.com/Altoros/cf-redislabs-broker.git cf-redislabs-broker
cd cf-redislabs-broker
./bin/build
```

After that you'll find resulting binary in `out/redislabs-service-broker`.

### How to run

```
./out/redislabs-service-broker -c config.yml
```

You can find an sample of config file in `examples` [folder](https://github.com/Altoros/cf-redislabs-broker/tree/master/examples).

### Adding a dependency

* Install [glide](https://github.com/Masterminds/glide.git)
* `glide get github.com/<ORG>/<REPO>`
* `find vendor -name '.git' | xargs rm -r`
* `git add vendor glide.yaml` & commit

Note: use `glide` with at least `0.7.2` version.

### Tests

```
go test ./redislabs/...
```

### How to run

```
go build -o redislabs-service-broker ./cmd
```
