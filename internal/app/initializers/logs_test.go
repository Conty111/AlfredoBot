package initializers_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Conty111/AlfredoBot/internal/app/initializers"
)

var _ = Describe("Logs", func() {
	Describe("InitializeLogs() ", func() {
		It("should initialize logs", func() {
			err := initializers.InitializeLogs()

			Expect(err).To(BeNil())
		})
	})
})

