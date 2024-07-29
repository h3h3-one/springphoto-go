package controllers

import (
	"database/sql"
	"github.com/gofiber/fiber/v2"
	"log/slog"
	"springoff/internal/config"
	"springoff/internal/models/album"
	"springoff/internal/models/upload"
	"strings"
)

type Album struct {
	album *album.Album
}

func AlbumNew(db *sql.DB) *Album {
	alb := album.New(db)
	return &Album{album: alb}
}

func (a *Album) Get(config *config.Config) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		idAlbum := c.Params("idAlbum")
		albumImages, err := a.album.GetImages(idAlbum, config)
		if err != nil {
			slog.Error("Error get album images", "error", err, "ID album", idAlbum)
			return c.SendStatus(500)
		}

		title, err := a.album.GetTitle(idAlbum)
		if err != err {
			slog.Error("Error get title album", "error", err, "id album", idAlbum)
			return c.SendStatus(500)
		}

		return c.Render("album",
			fiber.Map{"AlbumImages": albumImages,
				"Title": title})
	}
}

func (a *Album) Upload(config *config.Config) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		if form, err := c.MultipartForm(); err == nil {
			up, err := upload.New(form.Value["title"], form.File["cover"], form.File["albumImage"])
			if err != nil {
				slog.Error("Error validate forms adding album", "err", err)
				return c.Status(400).JSON(fiber.Map{
					"error": err.Error(),
				})
			}

			if err := a.album.Upload(up, config); err != nil {
				slog.Error("Error upload album", "err", err)
				return c.Status(400).JSON(fiber.Map{
					"error": err.Error(),
				})
			}
		}

		slog.Info("Album added successfully")
		return c.Status(200).JSON(fiber.Map{
			"successfully": "album added successfully",
		})
	}
}

func (a *Album) Delete(config *config.Config) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		temp := string(c.Body())               //id=...
		idAlbum := strings.Split(temp, "=")[1] // int id
		if err := a.album.Delete(idAlbum, config); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(200).JSON(fiber.Map{
			"successfully": "service deleted successfully",
		})
	}
}

type SwapRequest struct {
	Id    int    `json:"id"`
	Shift string `json:"shift"`
}

func (a *Album) Swap() func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		req := new(SwapRequest)
		if err := c.BodyParser(&req); err != nil {
			slog.Error("Error parse body", "error", err)
			return c.Status(500).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		if err := a.album.Swap(req.Id, req.Shift); err != nil {
			slog.Error("Error swap service", "error", err)
			return c.Status(500).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(200).JSON(fiber.Map{
			"successfully": "service swap successfully",
		})
	}
}
