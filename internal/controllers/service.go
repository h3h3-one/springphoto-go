package controllers

import (
	"database/sql"
	"github.com/gofiber/fiber/v2"
	"log/slog"
	"springoff/internal/config"
	"springoff/internal/models/service"
	"strings"
)

type Service struct {
	service *service.Service
}

func ServiceNew(db *sql.DB) *Service {
	srv := service.New(db)
	return &Service{service: srv}
}

func (s *Service) Get() func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		services, err := s.service.GetServices()
		if err != nil {
			slog.Error("Error get services", "error", err)
			return c.SendStatus(500)
		}
		return c.Render("service", fiber.Map{"Title": "Прайс", "Service": "active", "Services": services})
	}
}

func (s *Service) Upload(config *config.Config) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		if form, err := c.MultipartForm(); err == nil {

			err := s.service.UploadService(form, config)
			if err != nil {
				return c.Status(500).JSON(fiber.Map{
					"error": err.Error(),
				})
			}
			return c.Redirect("/service")
		}
		return c.Status(500).JSON(fiber.Map{
			"error": "service not upload",
		})
	}
}

func (s *Service) Swap() func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		req := new(SwapRequest)
		if err := c.BodyParser(&req); err != nil {
			slog.Error("Error parse body", "error", err)
			return c.Status(500).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		if err := s.service.Swap(req.Id, req.Shift); err != nil {
			slog.Error("Error swap album", "error", err)
			return c.Status(500).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(200).JSON(fiber.Map{
			"successfully": "album swap successfully",
		})
	}
}

func (s *Service) Delete() func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		temp := string(c.Body())               //id=...
		idAlbum := strings.Split(temp, "=")[1] // int id
		if err := s.service.Delete(idAlbum); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(200).JSON(fiber.Map{
			"successfully": "album deleted successfully",
		})
	}
}
