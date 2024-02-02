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

	Describe("GetLastIndex", func() {
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
				Expect(sm.GetLastIndex()).To(Equal(uint64(math.MaxUint64)))
			})
		})

		Context("when there is one element added", func() {
			BeforeEach(func() {
				err := sm.Append(3, []byte{1})
				Expect(err).ToNot(HaveOccurred())
			})
			It("should return the index of that element", func() {
				Expect(sm.GetLastIndex()).To(Equal(uint64(3)))
			})
		})

	})

	Describe("GetFirstIndex", func() {
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
				Expect(sm.GetFirstIndex()).To(Equal(uint64(math.MaxUint64)))
			})
		})

		Context("when there is one element added", func() {
			BeforeEach(func() {
				err := sm.Append(3, []byte{1})
				Expect(err).ToNot(HaveOccurred())
			})
			It("should return the index of that element", func() {
				Expect(sm.GetFirstIndex()).To(Equal(uint64(3)))
			})
		})

	})

	Describe("Count", func() {
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
			It("should return 0", func() {
				Expect(sm.Count()).To(Equal(uint64(0)))
			})
		})

		Context("when there is one element added", func() {
			BeforeEach(func() {
				err := sm.Append(3, []byte{1})
				Expect(err).ToNot(HaveOccurred())
			})
			It("should return the index of that element", func() {
				Expect(sm.Count()).To(Equal(uint64(1)))
			})
		})

	})

	Describe("Truncate", func() {
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
			Context("when I truncate", func() {
				var err error
				BeforeEach(func() {
					err = sm.Truncate()
				})

				It("should not return an error", func() {
					Expect(err).ToNot(HaveOccurred())
				})

				It("should have count 0", func() {
					Expect(sm.Count()).To(Equal(uint64(0)))
				})

				Context("when I add one element", func() {
					BeforeEach(func() {
						err := sm.Append(3, []byte{1})
						Expect(err).ToNot(HaveOccurred())
					})
					It("should have count 1", func() {
						Expect(sm.Count()).To(Equal(uint64(1)))
					})
				})

			})
		})

		Context("when there is one element", func() {
			BeforeEach(func() {
				err := sm.Append(3, []byte{1})
				Expect(err).ToNot(HaveOccurred())
			})
			Context("when I truncate", func() {
				var err error
				BeforeEach(func() {
					err = sm.Truncate()
				})

				It("should not return an error", func() {
					Expect(err).ToNot(HaveOccurred())
				})

				It("should have count 0", func() {
					Expect(sm.Count()).To(Equal(uint64(1)))
				})

				Context("when I add one element", func() {
					BeforeEach(func() {
						err := sm.Append(4, []byte{2})
						Expect(err).ToNot(HaveOccurred())
					})
					It("should have count 2", func() {
						Expect(sm.Count()).To(Equal(uint64(2)))
					})
					It("should contain data of the second element", func() {
						var d []byte
						err := sm.Read(4, func(data []byte) error {
							d = make([]byte, len(data))
							copy(d, data)
							return nil
						})

						Expect(err).ToNot(HaveOccurred())
						Expect(d).To(Equal([]byte{2}))
					})
					It("should contain data of the first element", func() {
						var d []byte
						err := sm.Read(3, func(data []byte) error {
							d = make([]byte, len(data))
							copy(d, data)
							return nil
						})

						Expect(err).ToNot(HaveOccurred())
						Expect(d).To(Equal([]byte{1}))
					})

				})

			})

		})

	})

	Describe("MaxSize", func() {
		When("I set MaxSize to 2k", func() {
			var sm *statemate.StateMate[uint64]
			BeforeEach(func() {
				var err error
				sm, err = statemate.Open[uint64](filepath.Join(tempDir, "state"), statemate.Options{
					MaxSize: 2048,
				})
				Expect(err).ToNot(HaveOccurred())
				DeferCleanup(func() {
					err := sm.Close()
					Expect(err).ToNot(HaveOccurred())
				})
			})

			When("I add 1k of data", func() {
				var err error
				BeforeEach(func() {
					err = sm.Append(1, make([]byte, 1024))
				})
				It("should not return an error", func() {
					Expect(err).ToNot(HaveOccurred())
				})

				When("I add 1k and one byte of data", func() {
					var err error
					BeforeEach(func() {
						err = sm.Append(2, make([]byte, 1025))
					})
					It("should return an error", func() {
						Expect(err).To(MatchError(statemate.ErrNotEnoughSpace))
					})
				})

				When("I add 1k of data", func() {
					var err error
					BeforeEach(func() {
						err = sm.Append(2, make([]byte, 1024))
					})
					It("should not return an error", func() {
						Expect(err).ToNot(HaveOccurred())
					})
				})

				When("I add 1k minus one byte of data", func() {
					var err error
					BeforeEach(func() {
						err = sm.Append(2, make([]byte, 1023))
					})
					It("should not return an error", func() {
						Expect(err).ToNot(HaveOccurred())
					})
				})

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
