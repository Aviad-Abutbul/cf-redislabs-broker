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

* Clone the repo:

```
git clone https://github.com/Altoros/redislabs-service-broker.git
```

* Execute `source .envrc` or `direnv allow` if you have [direnv](http://direnv.net/)

### Adding a dependency

* Install [glide](https://github.com/Masterminds/glide.git)
* `glide get github.com/<ORG>/<REPO>`
* `find vendor -name '.git' | xargs rm -r`
* `git add vendor glide.yaml` & commit
