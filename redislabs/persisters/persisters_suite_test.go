package persisters_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPersisters(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Persisters Suite")
}
