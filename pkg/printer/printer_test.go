package printer_test

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/jimmidyson/prettyconf/pkg/printer"
	"github.com/jimmidyson/prettyconf/pkg/printer/testdata"
)

var _ = Describe("Printer", func() {

	It("should properly print config with comments", func() {
		desiredConfig, err := ioutil.ReadFile(filepath.Join("testdata", "printed_toplevel.yaml"))
		Expect(err).NotTo(HaveOccurred())

		w := &bytes.Buffer{}
		Expect(printer.PrettyPrint(
			testdata.TopLevel{
				A: testdata.AStruct{
					D: 5,
					E: testdata.NestedStruct{
						F: "somestring",
					},
				},
				C: testdata.CStruct{
					H: "something new",
				},
			},
			w, logger)).To(Succeed())
		GinkgoWriter.Write(w.Bytes())
		Expect(strings.TrimSpace(w.String())).To(Equal(strings.TrimSpace(string(desiredConfig))))
	})
})
