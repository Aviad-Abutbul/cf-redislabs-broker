package redislabs_test

import (
	"github.com/altoros/redislabs-service-broker/redislabs"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Broker", func() {
	var (
		broker redislabs.ServiceBroker
		config redislabs.Config
	)

	JustBeforeEach(func() {
		broker = redislabs.ServiceBroker{
			Config: config,
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
})
