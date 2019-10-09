package astutils_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestASTUtils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AST utils Suite")
}
