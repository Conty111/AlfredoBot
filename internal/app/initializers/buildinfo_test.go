package initializers_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Conty111/TelegramBotTemplate/internal/app/initializers"
)

var _ = Describe("Buildinfo", func() {
	Describe("InitializeBuildInfo", func() {
		It("should initialize and return build.Info", func() {
			info := initializers.InitializeBuildInfo()

			Expect(info).NotTo(BeNil())
		})
	})
})
