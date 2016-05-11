package config_test

import (
	brokerconfig "github.com/RedisLabs/cf-redislabs-broker/redislabs/config"

	// "os"
	"path"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {

	var (
		config         brokerconfig.Config
		configPath     string
		parseConfigErr error
	)

	BeforeEach(func() {
		configPath = "valid_config.yml"
	})

	JustBeforeEach(func() {
		path, err := filepath.Abs(path.Join("assets", configPath))
		Ω(err).ToNot(HaveOccurred())
		config, parseConfigErr = brokerconfig.LoadFromFile(path)
	})

	Context("when the configuration is valid", func() {
		It("does not fail", func() {
			Ω(parseConfigErr).NotTo(HaveOccurred())
		})
		It("loads service broker name", func() {
			Ω(config.ServiceBroker.Name).To(Equal("my-redis"))
		})
		It("loads service id", func() {
			Ω(config.ServiceBroker.ServiceID).To(Equal("redislabs-service-broker-0b814f"))
		})
		It("loads service metadata", func() {
			Ω(config.ServiceBroker.Metadata.DisplayName).To(Equal("RedisLabs Enterprise Cluster"))
			Ω(config.ServiceBroker.Metadata.Image).To(Equal("base-64-image"))
			Ω(config.ServiceBroker.Metadata.ProviderDisplayName).To(Equal("RedisLabs"))
		})
		It("loads service broker plans", func() {
			Ω(config.ServiceBroker.Plans).To(Equal("redislabs-service-broker-0b814f"))
		})
	})

	Context("when the configuration file is not found", func() {
		BeforeEach(func() {
			configPath = "nonexistent_config.yml"
		})

		It("returns an error", func() {
			Ω(parseConfigErr).Should(MatchError(ContainSubstring("not found")))
		})
	})

	Context("when the configuration is invalid", func() {
		BeforeEach(func() {
			configPath = "invalid_config.yml"
		})
		It("fails", func() {
			Ω(parseConfigErr).To(HaveOccurred())
		})
	})

})
