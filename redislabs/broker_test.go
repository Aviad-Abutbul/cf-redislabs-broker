package redislabs_test

import (
	"github.com/altoros/redislabs-service-broker/redislabs"
	"github.com/pivotal-cf/brokerapi"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Broker", func() {
	var (
		broker   redislabs.ServiceBroker
		services []brokerapi.Service
	)

	BeforeEach(func() {
		broker = redislabs.ServiceBroker{}
		services = broker.Services()
	})

	Describe("Looking for plans", func() {
		It("Offers a service with at least one plan to use", func() {
			Expect(len(services)).To(Equal(1))
			Expect(len(services[0].Plans)).ToNot(Equal(0))
		})
	})
})