package persisters_test

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/Altoros/cf-redislabs-broker/redislabs/persisters"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Persisters", func() {
	var (
		state persisters.State
	)

	BeforeEach(func() {
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
		var (
			local       = persisters.Local{}
			tmpStateDir string
		)

		Context("Given an instance state", func() {
			It("Appears to save it successfully and then loads it back", func() {
				tmpStateDir, _ = ioutil.TempDir("", "redislabs-state-test")
				defer os.RemoveAll(tmpStateDir)
				persisters.GetStatePath = func() (string, error) {
					return path.Join(tmpStateDir, "state.json"), nil
				}

				err := local.Save(&state)
				Expect(err).NotTo(HaveOccurred())
				loaded, err := local.Load()
				Expect(err).NotTo(HaveOccurred())
				Expect(loaded).To(Equal(&state))
			})
		})
	})
})
