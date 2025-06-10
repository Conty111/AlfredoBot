package build_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Conty111/AlfredoBot/internal/app/build"
)

var _ = Describe("Info", func() {
	Describe("NewInfo()", func() {
		It("should create new info object", func() {
			info := build.NewInfo()

			Expect(info).NotTo(BeNil())
		})
	})
})
