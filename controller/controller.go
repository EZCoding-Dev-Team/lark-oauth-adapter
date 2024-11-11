package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"lark-oauth-adapter/dto"
	"lark-oauth-adapter/service"
)

type Controller struct {
	authService *service.AuthService
}

type TokenRequest struct {
	GrantType    string `json:"grant_type"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func NewController() *Controller {
	return &Controller{
		authService: service.NewAuthService(),
	}
}

func (c *Controller) RegisterRoutes(app *fiber.App) {
	app.Post("token", c.accessToken)
	app.Get("userinfo", c.userInfo)
}

func (c *Controller) accessToken(ctx *fiber.Ctx) error {
	var data dto.AccessTokenRequest
	if err := ctx.BodyParser(&data); err != nil {
		log.Error(err)
		return ctx.SendStatus(fiber.StatusBadRequest)
	}
	resp, err := c.authService.GetAccessToken(data)
	if err != nil {
		log.Error(err)
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}
	return ctx.JSON(resp)
}

func (c *Controller) userInfo(ctx *fiber.Ctx) error {
	authHeader := ctx.Get("Authorization")
	if authHeader == "" {
		return ctx.SendStatus(fiber.StatusUnauthorized)
	}
	resp, err := c.authService.GetUserInfo(authHeader)
	if err != nil {
		log.Error(err)
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}
	return ctx.JSON(resp)
}
