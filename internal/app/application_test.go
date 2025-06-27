package app_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Conty111/AlfredoBot/internal/app"
	"github.com/Conty111/AlfredoBot/internal/configs"
)

var _ = Describe("Application", func() {
	Describe("BuildApplication()", func() {
		It("should create new Application", func() {
			app, err := app.BuildApplication(&configs.Configuration{
				App: &configs.App{},
				Telegram: &configs.TelegramConfig{
					Token: "test-token",
				},
			})

			Expect(app).NotTo(BeNil())
			Expect(err).To(BeNil())
		})
	})
})
