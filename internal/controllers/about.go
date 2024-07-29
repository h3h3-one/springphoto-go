package controllers

import (
	"database/sql"
	"github.com/gofiber/fiber/v2"
	"springoff/internal/models/about"
)

type About struct {
	about *about.About
}

func AboutNew(db *sql.DB) *About {
	abt := about.New(db)
	return &About{about: abt}
}

func (a *About) Get() func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		return c.Render("about", fiber.Map{"Title": "О себе", "About": "active"})
	}
}
