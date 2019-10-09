package loader_test

import (
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/jimmidyson/prettyconf/pkg/testutils"
)

func TestLoader(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Loader Suite")
}

var logger logr.Logger

var _ = BeforeEach(func() {
	logger = &testutils.GinkgoLogger{Writer: GinkgoWriter}
})
