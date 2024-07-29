package auth

import (
	"database/sql"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"springoff/internal/models/login"
)

type Uuid struct {
	Uuid string `cookie:"uuid"`
}

func Check(db *sql.DB) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		l := login.New(db)

		newUuid := new(Uuid)
		if err := ctx.CookieParser(newUuid); err != nil {
			return err
		}

		isAuth := l.UuidExist(newUuid.Uuid)

		err := ctx.Bind(fiber.Map{
			"Auth": isAuth,
		})
		if err != nil {
			return fmt.Errorf("error bind auth: %w", err)
		}
		return ctx.Next()
	}
}
