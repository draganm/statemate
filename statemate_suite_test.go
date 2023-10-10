package statemate_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/draganm/statemate"
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
)

func TestStatemate(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Statemate Suite", types.ReporterConfig{NoColor: true})
}

var _ = Describe("Statemate", func() {

	var tempDir string
	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "")
		Expect(err).ToNot(HaveOccurred())
		DeferCleanup(func() {
			err := os.RemoveAll(tempDir)
			Expect(err).ToNot(HaveOccurred())
		})

	})

	Describe("Open", func() {
		var err error
		var sm *statemate.StateMate
		BeforeEach(func() {
			sm, err = statemate.Open(filepath.Join(tempDir, "state"))
		})

		It("should not return an error", func() {
			Expect(err).ToNot(HaveOccurred())
		})

		It("should return an instance of statemate", func() {
			Expect(sm).ToNot(BeNil())
		})

	})
})
