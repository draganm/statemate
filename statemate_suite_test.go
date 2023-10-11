package statemate_test

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/draganm/statemate"
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
)

func TestStateMate(t *testing.T) {
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
		var sm *statemate.StateMate[uint64]
		BeforeEach(func() {
			sm, err = statemate.Open[uint64](filepath.Join(tempDir, "state"), statemate.Options{})
			if err != nil {
				DeferCleanup(func() {
					err := sm.Close()
					Expect(err).ToNot(HaveOccurred())
				})
			}
		})

		It("should not return an error", func() {
			Expect(err).ToNot(HaveOccurred())
		})

		It("should return an instance of statemate", func() {
			if err == nil {
				Expect(sm).ToNot(BeNil())
			}
		})
	})

	Describe("IsEmpty", func() {
		var sm *statemate.StateMate[uint64]
		BeforeEach(func() {
			var err error
			sm, err = statemate.Open[uint64](filepath.Join(tempDir, "state"), statemate.Options{})
			Expect(err).ToNot(HaveOccurred())
			DeferCleanup(func() {
				err := sm.Close()
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when statemate is empty", func() {
			It("should return true", func() {
				Expect(sm.IsEmpty()).To(BeTrue())
			})
		})

		Context("when there is one element added", func() {
			BeforeEach(func() {
				err := sm.Append(3, []byte{1})
				Expect(err).ToNot(HaveOccurred())
			})
			It("should return false", func() {
				Expect(sm.IsEmpty()).To(BeFalse())
			})
		})

	})

	Describe("LastIndex", func() {
		var sm *statemate.StateMate[uint64]
		BeforeEach(func() {
			var err error
			sm, err = statemate.Open[uint64](filepath.Join(tempDir, "state"), statemate.Options{})
			Expect(err).ToNot(HaveOccurred())
			DeferCleanup(func() {
				err := sm.Close()
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when statemate is empty", func() {
			It("should return math.MaxUint64", func() {
				Expect(sm.LastIndex()).To(Equal(uint64(math.MaxUint64)))
			})
		})

		Context("when there is one element added", func() {
			BeforeEach(func() {
				err := sm.Append(3, []byte{1})
				Expect(err).ToNot(HaveOccurred())
			})
			It("should return the index of that element", func() {
				Expect(sm.LastIndex()).To(Equal(uint64(3)))
			})
		})

	})

	Describe("adding data", func() {

		var sm *statemate.StateMate[uint64]
		BeforeEach(func() {
			var err error
			sm, err = statemate.Open[uint64](filepath.Join(tempDir, "state"), statemate.Options{})
			Expect(err).ToNot(HaveOccurred())
			DeferCleanup(func() {
				err := sm.Close()
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when statemate is empty", func() {
			Context("when I try to access any index", func() {
				var err error
				BeforeEach(func() {
					err = sm.Read(2, func(d []byte) error {
						return nil
					})
				})
				It("should return an error", func() {
					Expect(err).To(Equal(statemate.ErrNotFound))
				})
			})
			Context("when I Append() some data", func() {
				var err error
				BeforeEach(func() {
					err = sm.Append(1, []byte{1, 2, 3})
				})
				It("should not return an error", func() {
					Expect(err).ToNot(HaveOccurred())
				})

				Context("when I read the written data", func() {
					var err error
					var data []byte
					BeforeEach(func() {
						err = sm.Read(1, func(d []byte) error {
							data = make([]byte, len(d))
							copy(data, d)
							return nil
						})
					})
					It("should not return an error", func() {
						Expect(err).ToNot(HaveOccurred())
					})
					It("should read the written data", func() {
						Expect(data).To(Equal([]byte{1, 2, 3}))
					})
				})

				Context("when I try to add another chunk of data with lower index", func() {
					var err error
					BeforeEach(func() {
						err = sm.Append(1, []byte{4})
					})
					It("should return an error", func() {
						Expect(err).To(Equal(statemate.ErrIndexMustBeIncreasing))
					})
				})

				Context("when I try to add another chunk of data with an gap index", func() {
					var err error
					BeforeEach(func() {
						err = sm.Append(3, []byte{4})
					})
					It("should return an error", func() {
						Expect(err).To(Equal(statemate.ErrIndexGapsAreNotAllowed))
					})
				})

				Context("when I add another chunk of data", func() {
					var err error
					BeforeEach(func() {
						err = sm.Append(2, []byte{4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20})
					})
					It("should not return an error", func() {
						Expect(err).ToNot(HaveOccurred())
					})

					Context("when I read the second data chunk", func() {
						var err error
						var data []byte
						BeforeEach(func() {
							err = sm.Read(2, func(d []byte) error {
								data = make([]byte, len(d))
								copy(data, d)
								return nil
							})
						})
						It("should not return an error", func() {
							Expect(err).ToNot(HaveOccurred())
						})
						It("should read the written data", func() {
							Expect(data).To(Equal([]byte{4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}))
						})
					})
				})

			})
		})
	})
})
