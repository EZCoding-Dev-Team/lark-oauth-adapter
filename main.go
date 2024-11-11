package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"lark-oauth-adapter/controller"
)

func main() {
	ctrl := controller.NewController()

	app := fiber.New()
	ctrl.RegisterRoutes(app)

	err := app.Listen(":3000")
	if err != nil {
		log.Fatal(err)
	}
}
