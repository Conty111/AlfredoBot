package app_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Conty111/AlfredoBot/internal/app"
)

var _ = Describe("Application", func() {
	Describe("InitializeApplication()", func() {
		It("should create new Application", func() {
			app, err := app.InitializeApplication()

			Expect(app).NotTo(BeNil())
			Expect(err).To(BeNil())
		})
	})

	Describe("methods", func() {
		var (
			app *app.Application
		)

		BeforeEach(func() {
			app, _ = app.InitializeApplication()
		})

		Describe("Start(), Stop()", func() {
			It("should start and stop application", func() {
				ctx, cancel := context.WithCancel(context.Background())
				app.Start(ctx, false)

				defer cancel()

				err := app.Stop()

				Expect(err).To(BeNil())
			})
		})
	})
})
