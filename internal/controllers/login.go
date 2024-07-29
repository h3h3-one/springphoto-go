package controllers

import (
	"database/sql"
	"github.com/gofiber/fiber/v2"
	"log/slog"
	"springoff/internal/models/login"
)

type Login struct {
	login *login.Login
}

type Cookies struct {
	Uuid string `cookie:"uuid"`
}

func LoginNew(db *sql.DB) *Login {
	lg := login.New(db)
	return &Login{login: lg}
}

func (l *Login) NewAuthorization() func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		c := new(Cookies)
		if err := ctx.CookieParser(l); err != nil {
			slog.Error("error cookie parse uuid", "error", err)
			return err
		}
		slog.Info("parse cookie with uuid", "uuid", c.Uuid)
		id, err := l.login.NewId(c.Uuid)
		if err != nil {
			return err
		}
		if len(id) > 0 {
			cookie := new(fiber.Cookie)
			cookie.Name = "uuid"
			cookie.Value = id
			ctx.Cookie(cookie)
			slog.Info("send cookies to the client", "name", cookie.Name, "value", cookie.Value)
		}
		return ctx.Redirect("//localhost:8080/")
	}
}
