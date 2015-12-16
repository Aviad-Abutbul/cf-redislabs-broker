package redislabs_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path"

	"github.com/Altoros/cf-redislabs-broker/redislabs"
	"github.com/Altoros/cf-redislabs-broker/redislabs/cluster"
	brokerconfig "github.com/Altoros/cf-redislabs-broker/redislabs/config"
	"github.com/Altoros/cf-redislabs-broker/redislabs/instance_binders"
	"github.com/Altoros/cf-redislabs-broker/redislabs/instance_creators"
	"github.com/Altoros/cf-redislabs-broker/redislabs/persisters"
	"github.com/Altoros/cf-redislabs-broker/redislabs/testing"
	"github.com/ldmberman/brokerapi"
	"github.com/pivotal-golang/lager"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Broker", func() {
	var (
		broker    brokerapi.ServiceBroker
		config    brokerconfig.Config
		persister persisters.StatePersister
		logger    = lager.NewLogger("test") // does not actually log anything
	)

	JustBeforeEach(func() {
		broker = redislabs.NewServiceBroker(
			instancecreators.NewDefault(config, logger),
			instancebinders.NewDefault(config, logger),
			persister,
			config,
			logger,
		)
	})

	Describe("Looking for plans", func() {
		Context("Given a config with one default plan", func() {
			BeforeEach(func() {
				config = brokerconfig.Config{
					ServiceBroker: brokerconfig.ServiceBrokerConfig{
						Plans: []brokerconfig.ServicePlanConfig{
							{
								ID:          "",
								Name:        "",
								Description: "",
							},
						},
					},
				}
			})
			It("Offers a service with at least one plan to use", func() {
				Expect(len(broker.Services())).To(Equal(1))
				Expect(len(broker.Services()[0].Plans)).ToNot(Equal(0))
			})
		})
	})

	Describe("Provisioning an instance", func() {
		var (
			serviceID                           = "test-service-id"
			planID                              = "test-plan-id"
			requestedServiceID, requestedPlanID string
			params                              map[string]interface{}
			details                             brokerapi.ProvisionDetails
		)
		Context("Given a config with a default plan", func() {
			BeforeEach(func() {
				config = brokerconfig.Config{
					ServiceBroker: brokerconfig.ServiceBrokerConfig{
						ServiceID: serviceID,
						Plans: []brokerconfig.ServicePlanConfig{
							{
								ID:          planID,
								Name:        "test",
								Description: "Lorem ipsum dolor sit amet",
							},
						},
					},
					Redislabs: brokerconfig.RedislabsConfig{
						Address: "",
					},
				}
			})
			JustBeforeEach(func() {
				details = brokerapi.ProvisionDetails{
					ID:               requestedServiceID,
					PlanID:           requestedPlanID,
					OrganizationGUID: "",
					SpaceGUID:        "",
					Parameters:       params,
				}
			})
			Context("And a wrong service ID", func() {
				BeforeEach(func() {
					requestedServiceID = "unknown"
				})
				It("Rejects to create an instance", func() {
					_, err := broker.Provision("some-id", details, false)
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(redislabs.ErrServiceDoesNotExist))
				})
			})

			Context("And a wrong plan ID", func() {
				BeforeEach(func() {
					requestedServiceID = serviceID
					requestedPlanID = "unknown"
				})
				It("Rejects to create an instance", func() {
					_, err := broker.Provision("some-id", details, false)
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(redislabs.ErrPlanDoesNotExist))
				})
			})

			Context("And no database name", func() {
				BeforeEach(func() {
					requestedPlanID = planID
				})
				It("Complains about the database name", func() {
					_, err := broker.Provision("some-id", details, false)
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(redislabs.ErrDatabaseNameIsRequired))
				})
			})

			Context("And an empty name for the database", func() {
				BeforeEach(func() {
					params = map[string]interface{}{
						"name": "",
					}
				})
				It("Complains about the database name", func() {
					_, err := broker.Provision("some-id", details, false)
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(redislabs.ErrDatabaseNameIsRequired))
				})
			})

			Context("And given proper settings", func() {
				var (
					tmpStateDir string
					proxy       testing.HTTPProxy
					err         error
					settings    cluster.InstanceSettings
				)

				BeforeEach(func() {
					params = map[string]interface{}{
						"name": "test",
					}
					tmpStateDir, err = ioutil.TempDir("", "redislabs-state-test")
					Expect(err).NotTo(HaveOccurred())
					persister = persisters.NewLocalPersister(path.Join(tmpStateDir, "state.json"))

					proxy = testing.NewHTTPProxy()
					proxy.RegisterEndpointHandler("/", func(w http.ResponseWriter, r *http.Request) interface{} {
						decoder := json.NewDecoder(r.Body)
						defer r.Body.Close()
						if err := decoder.Decode(&settings); err != nil {
							Expect(err).NotTo(HaveOccurred())
						}
						return map[string]interface{}{
							"uid": 1,
							"authentication_redis_pass": "pass",
							"endpoint_ip":               []string{"10.0.2.4"},
							"dns_address_master":        "domain.com:11909",
							"status":                    "active",
						}
					})
					config.Redislabs.Address = proxy.URL()

					config.ServiceBroker.Plans[0].ServiceInstanceConfig = brokerconfig.ServiceInstanceConfig{
						MemoryLimit: 1024,
						Replication: true,
						Persistence: "disabled",
					}
				})
				AfterEach(func() {
					proxy.Close()
					os.RemoveAll(tmpStateDir)
				})

				It("Creates an instance of the configured default plan", func() {
					_, err := broker.Provision("some-id", details, false)
					Expect(err).ToNot(HaveOccurred())
					Expect(settings.MemoryLimit).To(Equal(int64(1024)))
					Expect(settings.Replication).To(Equal(true))
					Expect(settings.Persistence).To(Equal("disabled"))
					Expect(settings.Sharding).To(Equal(false))
					Expect(settings.ImplicitShardKey).To(Equal(false))
				})
				It("Rejects to provision the same instance again", func() {
					broker.Provision("some-id", details, false)
					_, err := broker.Provision("some-id", details, false)
					Expect(err).To(HaveOccurred())
				})
				It("Saves the credentials properly", func() {
					_, err := broker.Provision("some-id", details, false)
					Expect(err).ToNot(HaveOccurred())

					state, err := persister.Load()
					Expect(err).ToNot(HaveOccurred())
					Expect(len(state.AvailableInstances)).To(Equal(1))
					s := state.AvailableInstances[0]
					Expect(s.ID).To(Equal("some-id"))
					Expect(s.Credentials).To(Equal(cluster.InstanceCredentials{
						UID:      1,
						Port:     11909,
						IPList:   []string{"10.0.2.4"},
						Password: "pass",
					}))
				})
				Context("And when requested for more than one shard", func() {
					BeforeEach(func() {
						config.ServiceBroker.Plans[0].ServiceInstanceConfig = brokerconfig.ServiceInstanceConfig{
							MemoryLimit: 2048,
							ShardCount:  2,
						}
					})
					It("Setups the sharding properly", func() {
						_, err := broker.Provision("some-id", details, false)
						Expect(err).NotTo(HaveOccurred())
						Expect(settings.MemoryLimit).To(Equal(int64(2048)))
						Expect(settings.ShardCount).To(Equal(int64(2)))
						Expect(settings.Sharding).To(Equal(true))
						Expect(settings.ImplicitShardKey).To(Equal(true))
						Expect(settings.ShardKeyRegex).To(Equal(map[string]string{
							`.*\{(?<tag>.*)\}.*`: "Hashing is done on the substring between the curly braces.",
							`(?<tag>.*)`:         "The entire key's name is used for hashing.",
						}))
					})
				})
				Context("And when requested for snapshots", func() {
					BeforeEach(func() {
						config.ServiceBroker.Plans[0].ServiceInstanceConfig = brokerconfig.ServiceInstanceConfig{
							Persistence: "snapshot",
							Snapshot: brokerconfig.Snapshot{
								Writes: 10,
								Secs:   12,
							},
						}
					})
					It("Applies the given snapshot configuration", func() {
						_, err := broker.Provision("some-id", details, false)
						Expect(err).NotTo(HaveOccurred())
						Expect(settings.Persistence).To(Equal("snapshot"))
						Expect(len(settings.Snapshot)).To(Equal(1))
						Expect(settings.Snapshot[0]).To(Equal(cluster.Snapshot{
							Writes: 10,
							Secs:   12,
						}))
					})
				})
			})
		})
	})

	Describe("Binding provisioned instances", func() {
		var (
			details brokerapi.BindDetails
		)
		BeforeEach(func() {
			details = brokerapi.BindDetails{
				AppGUID:   "",
				ServiceID: "test-service",
				PlanID:    "test-plan",
			}
		})
		Context("When there are no provisioned instances", func() {
			It("Rejects to bind anything", func() {
				_, err := broker.Bind("instance-id", "binding-id", details)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(brokerapi.ErrInstanceDoesNotExist))
			})
		})
		Context("When there is a provisioned instance", func() {
			var (
				tmpStateDir string
				state       *persisters.State
				err         error
			)
			BeforeEach(func() {
				tmpStateDir, err = ioutil.TempDir("", "redislabs-state-test")
				if err != nil {
					panic(err)
				}
				persister = persisters.NewLocalPersister(path.Join(tmpStateDir, "state.json"))
				state = &persisters.State{
					AvailableInstances: []persisters.ServiceInstance{
						{
							ID: "test-instance",
							Credentials: cluster.InstanceCredentials{
								UID:      1,
								Port:     11909,
								IPList:   []string{"10.0.2.5"},
								Password: "pass",
							},
						},
					},
				}
				if err = persister.Save(state); err != nil {
					panic(err)
				}
			})
			AfterEach(func() {
				os.RemoveAll(tmpStateDir)
			})
			It("Successfully retrieves the credentials", func() {
				credentials, err := broker.Bind("test-instance", "test-binding", details)
				Expect(err).NotTo(HaveOccurred())
				Expect(credentials).To(Equal(map[string]interface{}{
					"port":     11909,
					"ip_list":  []string{"10.0.2.5"},
					"password": "pass",
				}))
			})
		})
	})

	Describe("Deprovisioning instances", func() {
		var (
			err error
		)
		Context("When there are no provisioned instances", func() {
			It("Deprovisioning results in a failure", func() {
				err = broker.Deprovision("instance-id")
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(brokerapi.ErrInstanceDoesNotExist))
			})
		})
		Context("When a provisioned instance exists", func() {
			var (
				tmpStateDir string
				state       *persisters.State
				proxy       testing.HTTPProxy
			)
			BeforeEach(func() {
				tmpStateDir, err = ioutil.TempDir("", "redislabs-state-test")
				if err != nil {
					panic(err)
				}
				persister = persisters.NewLocalPersister(path.Join(tmpStateDir, "state.json"))
				state = &persisters.State{
					AvailableInstances: []persisters.ServiceInstance{
						{
							ID:          "test-instance",
							Credentials: cluster.InstanceCredentials{},
						},
					},
				}
				if err = persister.Save(state); err != nil {
					panic(err)
				}

				proxy = testing.NewHTTPProxy()
				proxy.RegisterEndpoints([]testing.Endpoint{{"/", ""}})
				config.Redislabs.Address = proxy.URL()
			})
			AfterEach(func() {
				proxy.Close()
				os.RemoveAll(tmpStateDir)
			})
			It("Can delete it successfully", func() {
				err = broker.Deprovision("test-instance")
				Expect(err).NotTo(HaveOccurred())
				err = broker.Deprovision("test-instance")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Updating instances", func() {
		Context("When the broker does not offer any services", func() {
			It("An update fails", func() {
				_, err := broker.Update("test-instance", brokerapi.UpdateDetails{}, false)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(redislabs.ErrServiceDoesNotExist))
			})
		})
		Context("When there are no instances", func() {
			BeforeEach(func() {
				config = brokerconfig.Config{
					ServiceBroker: brokerconfig.ServiceBrokerConfig{
						ServiceID: "test-service",
					},
				}
			})
			It("Fails", func() {
				_, err := broker.Update("test-instance", brokerapi.UpdateDetails{ID: "test-service"}, false)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(brokerapi.ErrInstanceDoesNotExist))
			})
		})
		Context("When there is an instance to update", func() {
			var (
				proxy       testing.HTTPProxy
				tmpStateDir string
				err         error

				updateSettings map[string]interface{}
			)
			BeforeEach(func() {
				tmpStateDir, err = ioutil.TempDir("", "redislabs-state-test")
				if err != nil {
					panic(err)
				}
				persister = persisters.NewLocalPersister(path.Join(tmpStateDir, "state.json"))

				proxy = testing.NewHTTPProxy()
				proxy.RegisterEndpoints([]testing.Endpoint{
					{"/v1/bdbs", map[string]interface{}{
						"uid": 1,
						"authentication_redis_pass": "pass",
						"endpoint_ip":               []string{"10.0.2.4"},
						"dns_address_master":        "domain.com:11909",
						"status":                    "active",
					}}})
				proxy.RegisterEndpointHandler("/v1/bdbs/1", func(w http.ResponseWriter, r *http.Request) interface{} {
					bytes, err := ioutil.ReadAll(r.Body)
					if err != nil {
						panic(err)
					}
					if err = json.Unmarshal(bytes, &updateSettings); err != nil {
						w.WriteHeader(422)
						return map[string]interface{}{
							"description": "invalid input data",
						}
					}
					return nil
				})

				config = brokerconfig.Config{
					ServiceBroker: brokerconfig.ServiceBrokerConfig{
						ServiceID: "test-service",
						Plans: []brokerconfig.ServicePlanConfig{
							{
								ID:          "test-plan-1",
								Name:        "test-1",
								Description: "",
								ServiceInstanceConfig: brokerconfig.ServiceInstanceConfig{
									MemoryLimit: 200000000,
									Replication: false,
									ShardCount:  1,
								},
							},
							{
								ID:   "test-plan-2",
								Name: "test-2",
								ServiceInstanceConfig: brokerconfig.ServiceInstanceConfig{
									MemoryLimit: 700000000,
									Replication: true,
									ShardCount:  2,
									Persistence: "snapshot",
									Snapshot: brokerconfig.Snapshot{
										Writes: 100,
										Secs:   10,
									},
								},
							},
						},
					},
					Redislabs: brokerconfig.RedislabsConfig{
						Address: proxy.URL(),
					},
				}
			})
			AfterEach(func() {
				proxy.Close()
				os.RemoveAll(tmpStateDir)
			})
			JustBeforeEach(func() {
				_, err = broker.Provision("test-instance", brokerapi.ProvisionDetails{
					ID:               "test-service",
					PlanID:           "test-plan-1",
					OrganizationGUID: "",
					SpaceGUID:        "",
					Parameters: map[string]interface{}{
						"name": "test",
					},
				}, false)
				if err != nil {
					panic(err)
				}
			})
			It("Updates its memory limit", func() {
				_, err = broker.Update("test-instance", brokerapi.UpdateDetails{
					ID: "test-service",
					Parameters: map[string]interface{}{
						"memory_size": 400000000,
					},
				}, false)
				Expect(err).NotTo(HaveOccurred())
				Expect(updateSettings).To(HaveKey("memory_size"))
				Expect(updateSettings["memory_size"]).To(BeEquivalentTo(400000000))
			})
			It("Updates its plan", func() {
				_, err = broker.Update("test-instance", brokerapi.UpdateDetails{
					ID:     "test-service",
					PlanID: "test-plan-2",
				}, false)
				Expect(err).NotTo(HaveOccurred())
				Expect(updateSettings).To(HaveKey("memory_size"))
				Expect(updateSettings["memory_size"]).To(BeEquivalentTo(700000000))
				Expect(updateSettings).To(HaveKey("replication"))
				Expect(updateSettings["replication"]).To(BeEquivalentTo(true))

				Expect(updateSettings).To(HaveKey("shards_count"))
				Expect(updateSettings["shards_count"]).To(BeEquivalentTo(2))
				Expect(updateSettings).To(HaveKey("sharding"))
				Expect(updateSettings["sharding"]).To(BeEquivalentTo(true))
				Expect(updateSettings).To(HaveKey("implicit_shard_key"))
				Expect(updateSettings["implicit_shard_key"]).To(BeEquivalentTo(true))
				Expect(updateSettings).To(HaveKey("shard_key_regex"))
				r := updateSettings["shard_key_regex"].(map[string]interface{})
				Expect(r[`.*\{(?<tag>.*)\}.*`]).To(BeEquivalentTo("Hashing is done on the substring between the curly braces."))
				Expect(r[`(?<tag>.*)`]).To(BeEquivalentTo("The entire key's name is used for hashing."))

				Expect(updateSettings).To(HaveKey("data_persistence"))
				Expect(updateSettings["data_persistence"]).To(BeEquivalentTo("snapshot"))
				Expect(updateSettings).To(HaveKey("snapshot_policy"))
				s := updateSettings["snapshot_policy"].([]interface{})[0].(map[string]interface{})
				Expect(s["secs"]).To(BeEquivalentTo(10))
				Expect(s["writes"]).To(BeEquivalentTo(100))
			})
			It("Updates both its plan and other parameters", func() {
				_, err = broker.Update("test-instance", brokerapi.UpdateDetails{
					ID:     "test-service",
					PlanID: "test-plan-2",
					Parameters: map[string]interface{}{
						"memory_size":      300000000,
						"data_persistence": "aof",
					},
				}, false)
				Expect(err).NotTo(HaveOccurred())
				Expect(updateSettings).To(HaveKey("memory_size"))
				Expect(updateSettings["memory_size"]).To(BeEquivalentTo(300000000))
				Expect(updateSettings).To(HaveKey("replication"))
				Expect(updateSettings["replication"]).To(BeEquivalentTo(true))
				Expect(updateSettings).To(HaveKey("shards_count"))
				Expect(updateSettings["shards_count"]).To(BeEquivalentTo(2))
				Expect(updateSettings).To(HaveKey("data_persistence"))
				Expect(updateSettings["data_persistence"]).To(BeEquivalentTo("aof"))
			})
			It("Rejects to update it to an unknown plan", func() {
				_, err = broker.Update("test-instance", brokerapi.UpdateDetails{
					ID:     "test-service",
					PlanID: "test-plan-3",
				}, false)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(redislabs.ErrPlanDoesNotExist))
			})
			It("Fails to process data of invalid type", func() {
				_, err = broker.Update("test-instance", brokerapi.UpdateDetails{
					ID: "test-service",
					Parameters: map[string]interface{}{
						"memory_size": "{{{",
					},
				}, false)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(cluster.ErrInvalidType("memory_size")))
			})
			It("Fails to process unknown properties", func() {
				_, err = broker.Update("test-instance", brokerapi.UpdateDetails{
					ID: "test-service",
					Parameters: map[string]interface{}{
						"unknown": 0,
					},
				}, false)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(cluster.ErrUnknownParam("unknown")))
			})
		})
	})

	Describe("Fetching the catalog", func() {
		Context("Given a config with a service with the ID, name, description, and plan", func() {
			BeforeEach(func() {
				config = brokerconfig.Config{
					ServiceBroker: brokerconfig.ServiceBrokerConfig{
						ServiceID:   "redislabs-test",
						Name:        "redislabs test",
						Description: "redislabs description",
						Plans: []brokerconfig.ServicePlanConfig{
							{
								ID:          "plan-1",
								Name:        "plan",
								Description: "plan description",
							},
						},
					},
				}
			})
			It("Provides them via a catalog request", func() {
				services := broker.Services()
				Expect(len(services)).To(Equal(1))

				service := services[0]
				Expect(service.ID).To(Equal("redislabs-test"))
				Expect(service.Name).To(Equal("redislabs test"))
				Expect(service.Description).To(Equal("redislabs description"))
				Expect(len(service.Plans)).To(Equal(1))

				plan := service.Plans[0]
				Expect(plan).To(Equal(brokerapi.ServicePlan{
					ID:          "plan-1",
					Name:        "plan",
					Description: "plan description",
				}))
			})
			It("Assigns a tag", func() {
				services := broker.Services()
				Expect(len(services)).To(Equal(1))
				service := services[0]

				Expect(len(service.Tags)).To(Equal(1))
				Expect(service.Tags[0]).To(Equal("redislabs"))
			})
			It("Says that it is bindable", func() {
				services := broker.Services()
				Expect(len(services)).To(Equal(1))
				service := services[0]

				Expect(service.Bindable).To(Equal(true))
			})
			It("Says that the plan is updatable", func() {
				services := broker.Services()
				Expect(len(services)).To(Equal(1))
				service := services[0]

				Expect(service.PlanUpdatable).To(Equal(true))
			})
		})
	})
})
