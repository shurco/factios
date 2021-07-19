package main

import (
	"embed"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/template/html"
	"github.com/shurco/factios/db"
	"github.com/shurco/factios/logger"
	"github.com/shurco/factios/model"
)

var (
	//go:embed template/*
	viewsfs embed.FS

	log  = logger.GetLogger("factios")
	base = db.NewDB("../db/")
)

func main() {
	engine := html.NewFileSystem(http.FS(viewsfs), ".html")

	app := fiber.New(fiber.Config{
		Views:                 engine,
		DisableStartupMessage: true,
		ServerHeader:          "factios",
	})
	app.Use(cors.New())
	app.Use(csrf.New())

	setupRoutes(app)

	err := app.Listen(":3000")
	if err != nil {
		log.Error().Err(err)
	}
}

func setupRoutes(app *fiber.App) {
	app.Get("/", factPage)
	app.Get("/f/", factPage)
	app.Get("/f/:lng/:short?", factPage)

	app.Get("/robots.txt", robotsFile)
	app.Get("/favicon.ico", faviconFile)

	app.Get("/api/:lng/fact", getRandomFact)
	app.Get("/api/:lng/fact/:short", getFactByID)

	app.Get("/ping", pingPong)

	app.Use(func(c *fiber.Ctx) error {
		return c.SendStatus(404)
	})
}

func pingPong(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "pong"})
}

func factPage(c *fiber.Ctx) error {
	t := time.Now()
	fact := &model.Fact{}
	var err error

	lng := c.Params("lng")
	if lng == "" || lng == "undefined" {
		lng = "ru"
	}

	shot := c.Params("short")
	if shot == "" || shot == "undefined" {
		fact, err = base.GetRandomFact(lng)
		if err != nil {
			return c.SendStatus(404)
		}
	} else {
		fact, err = base.GetFactByID(lng, shot)
		if err != nil {
			return c.SendStatus(404)
		}
	}

	return c.Render("template/index", fiber.Map{
		"Lang":  lng,
		"Fact":  fact.Fact,
		"Short": fact.Short,
		"Year":  t.Year(),
	})
}

func robotsFile(c *fiber.Ctx) error {
	return c.SendFile("./template/robots.txt")
}

func faviconFile(c *fiber.Ctx) error {
	return c.SendFile("./template/favicon.ico")
}

func getRandomFact(c *fiber.Ctx) error {
	fact, err := base.GetRandomFact(c.Params("lng"))
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fact)
}

func getFactByID(c *fiber.Ctx) error {
	fact, err := base.GetFactByID(c.Params("lng"), c.Params("short"))
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fact)
}
