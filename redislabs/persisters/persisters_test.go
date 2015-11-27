package persisters_test

import (
	"github.com/Altoros/cf-redislabs-broker/redislabs/persisters"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Persisters", func() {
	var (
		local persisters.Local
		state persisters.State
	)

	BeforeEach(func() {
		local = persisters.Local{}
		state = persisters.State{
			AvailableInstances: []persisters.ServiceInstance{
				{
					ID:       "test-id",
					Port:     11909,
					IPList:   []string{"10.0.0.1", "10.0.0.2"},
					Password: "passw0rd",
				},
			},
		}
	})

	Describe("Local JSON file persister", func() {
		Context("Given an instance state", func() {
			It("Appears to save it successfully", func() {
				err := local.Save(&state)
				Expect(err).NotTo(HaveOccurred())
			})
			It("Loads the previously saved state", func() {
				loaded, err := local.Load()
				Expect(err).NotTo(HaveOccurred())
				Expect(loaded).To(Equal(state))
			})
		})
	})
})
