package redislabs_test

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/Altoros/cf-redislabs-broker/redislabs"
	"github.com/Altoros/cf-redislabs-broker/redislabs/cluster"
	brokerconfig "github.com/Altoros/cf-redislabs-broker/redislabs/config"
	"github.com/Altoros/cf-redislabs-broker/redislabs/instance_binders"
	"github.com/Altoros/cf-redislabs-broker/redislabs/instance_creators"
	"github.com/Altoros/cf-redislabs-broker/redislabs/persisters"
	"github.com/Altoros/cf-redislabs-broker/redislabs/testing"
	"github.com/pivotal-cf/brokerapi"
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
				}
			})
			Context("And a wrong service ID", func() {
				BeforeEach(func() {
					requestedServiceID = "unknown"
				})
				It("Rejects to create an instance", func() {
					err := broker.Provision("some-id", details)
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
					err := broker.Provision("some-id", details)
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(redislabs.ErrPlanDoesNotExist))
				})
			})

			Context("And given proper settings", func() {
				var (
					tmpStateDir string
					proxy       testing.HTTPProxy
					err         error
				)

				BeforeEach(func() {
					requestedPlanID = planID

					tmpStateDir, err = ioutil.TempDir("", "redislabs-state-test")
					Expect(err).NotTo(HaveOccurred())
					persister = persisters.NewLocalPersister(path.Join(tmpStateDir, "state.json"))

					proxy = testing.NewHTTPProxy()
					proxy.RegisterEndpoints([]testing.Endpoint{
						{"/", map[string]interface{}{
							"uid": 1,
							"authentication_redis_pass": "pass",
							"endpoint_ip":               []string{"10.0.2.4"},
							"dns_address_master":        "domain.com:11909",
							"status":                    "active",
						}},
					})
					config.Redislabs.Address = proxy.URL()
				})
				AfterEach(func() {
					proxy.Close()
					os.RemoveAll(tmpStateDir)
				})

				It("Creates an instance of the configured default plan", func() {
					err := broker.Provision("some-id", details)
					Expect(err).ToNot(HaveOccurred())
				})
				It("Rejects to provision the same instance again", func() {
					broker.Provision("some-id", details)
					err := broker.Provision("some-id", details)
					Expect(err).To(HaveOccurred())
				})
				It("Saves the credentials properly", func() {
					err := broker.Provision("some-id", details)
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
})
