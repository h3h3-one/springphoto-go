package main

import (
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"
	"log"
	"log/slog"
	"springoff/internal/config"
	"springoff/internal/controllers"
	"springoff/internal/logger"
	"springoff/internal/middleware/auth"
	"springoff/internal/storage/sqlite3"
)

type Users struct {
	idUser       int
	loginUser    string
	passwordUser string
}

func main() {
	//configuration
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("Configuration fail init: %s", err)
	}

	//logger
	logger.New(cfg.Env)

	//database
	db, err := sqlite3.New(cfg.Storage)
	if err != nil {
		log.Fatalf("Database fail init: %s", err)
	}

	//template engine
	engine := html.New("/root/springoff/internal/views", ".html")
	app := fiber.New(fiber.Config{
		Views:       engine,
		ViewsLayout: "layouts/main",
		BodyLimit:   60 * 1024 * 1024,
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError

			var e *fiber.Error
			if errors.As(err, &e) {
				code = e.Code
			}
			return ctx.Status(code).Render("error", fiber.Map{"Title": fmt.Sprintf("Error %d", code), "Message": e.Message, "Code": code})
		},
	})
	//middleware
	app.Use(auth.Check(db))
	app.Use(recover.New())
	//static files
	app.Static("/", "/root/springoff/static", fiber.Static{Compress: true})

	//controllers
	home := controllers.HomeNew(db)
	service := controllers.ServiceNew(db)
	about := controllers.AboutNew(db)
	contact := controllers.ContactNew(db)
	album := controllers.AlbumNew(db)
	login := controllers.LoginNew(db)

	//routes
	app.Get("/", home.GetAll())
	app.Get("/service", service.Get()).Name("service")
	app.Get("/about", about.Get())
	app.Get("/contact", contact.Get())
	app.Get("/albums/:idAlbum", album.Get(cfg))

	api := app.Group("/admin", basicauth.New(basicauth.Config{
		Authorizer: func(login string, password string) bool {
			user := Users{}

			row := db.QueryRow("SELECT * FROM users WHERE login=?", login)
			err := row.Scan(&user.idUser, &user.loginUser, &user.passwordUser)
			if err != nil {
				slog.Error("fail mapping Users table", "error", err)
			}

			slog.Info("user authentication", "login", login, "password", password)
			if login == user.loginUser && password == user.passwordUser {
				slog.Info("user successfully logged in")
				return true
			}
			slog.Info("the user is not authenticated")
			return false
		},
	}))

	api.Use("/metrics", monitor.New())

	api.Get("/login", login.NewAuthorization())
	api.Get("/album/adding", func(ctx *fiber.Ctx) error {
		return ctx.Render("adding-album", fiber.Map{"Adding": "active", "Title": "Добавить альбом"})
	})

	api.Post("/album", album.Upload(cfg))
	api.Delete("/album", album.Delete(cfg))
	api.Put("/album/swap", album.Swap())

	api.Post("/service", service.Upload(cfg))
	api.Put("/service/swap", service.Swap())
	api.Delete("/service", service.Delete())

	log.Fatal(app.Listen(cfg.Address + cfg.Port))
}
