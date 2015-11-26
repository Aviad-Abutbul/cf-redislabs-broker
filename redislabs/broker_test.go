package redislabs_test

import (
	"github.com/altoros/redislabs-service-broker/redislabs"
	"github.com/altoros/redislabs-service-broker/redislabs/adapters"
	"github.com/pivotal-cf/brokerapi"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Broker", func() {
	var (
		broker          redislabs.ServiceBroker
		config          redislabs.Config
		instanceCreator redislabs.ServiceInstanceCreator
		instanceBinders redislabs.ServiceInstanceBinder
	)

	JustBeforeEach(func() {
		broker = redislabs.ServiceBroker{
			Config:          config,
			InstanceCreator: instanceCreator,
			InstanceBinder:  instanceBinder,
		}
	})

	Describe("Looking for plans", func() {
		Context("Given a config with one default plan", func() {
			BeforeEach(func() {
				config = redislabs.Config{
					DefaultPlans: []redislabs.ServicePlanConfig{
						{
							ID:          "",
							Name:        "test",
							Description: "",
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
			JustBeforeEach(func() {
				config = redislabs.Config{
					ServiceID: serviceID,
					DefaultPlans: []redislabs.ServicePlanConfig{
						{
							ID:          planID,
							Name:        "test",
							Description: "",
						},
					},
				}
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
					Expect(err).To(Equal(brokerapi.ErrInstanceDoesNotExist))
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
			Context("And no instance creators", func() {
				BeforeEach(func() {
					requestedServiceID = serviceID
					requestedPlanID = planID
				})
				It("Rejects to create an instance", func() {
					err := broker.Provision("some-id", details)
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(redislabs.ErrInstanceCreatorNotFound))
				})
			})
			Context("And given a default instance creator and correct plan and instance IDs", func() {
				BeforeEach(func() {
					requestedServiceID = serviceID
					requestedPlanID = planID
					instanceCreator = &adapters.DefaultCreator{}
				})
				It("Creates an instance of the configured default plan", func() {
					err := broker.Provision("some-id", details)
					Expect(err).ToNot(HaveOccurred())
				})
				It("Rejects to provision the same instance again", func() {
					err := broker.Provision("some-id", details)
					Expect(err).To(HaveOccurred())
				})
			})
		})
	})
})
