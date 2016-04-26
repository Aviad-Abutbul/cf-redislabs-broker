package persisters_test

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/RedisLabs/cf-redislabs-broker/redislabs/cluster"
	"github.com/RedisLabs/cf-redislabs-broker/redislabs/persisters"

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
					ID: "test-id",
					Credentials: cluster.InstanceCredentials{
						UID:      1,
						Port:     11909,
						IPList:   []string{"10.0.0.1", "10.0.0.2"},
						Password: "passw0rd",
					},
				},
			},
		}
	})

	Describe("Local JSON file persister", func() {
		Context("Given an instance state", func() {
			It("Appears to save it successfully and then loads it back", func() {
				tmpStateDir, err := ioutil.TempDir("", "redislabs-state-test")
				Expect(err).NotTo(HaveOccurred())
				defer os.RemoveAll(tmpStateDir)
				local := persisters.NewLocalPersister(path.Join(tmpStateDir, "state.json"))
				err = local.Save(&state)
				Expect(err).NotTo(HaveOccurred())
				loaded, err := local.Load()
				Expect(err).NotTo(HaveOccurred())
				Expect(loaded).To(Equal(&state))
			})
		})
	})
})
