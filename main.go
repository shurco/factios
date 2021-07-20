package main

import (
	"embed"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/template/html"
	"github.com/ikeikeikeike/go-sitemap-generator/v2/stm"

	"github.com/shurco/factios/database"
	"github.com/shurco/factios/logger"
	"github.com/shurco/factios/model"
)

var (
	//go:embed template/*
	viewsfs embed.FS

	log  = logger.GetLogger("server")
	base = database.NewDB("./db/")
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

	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))

	setupRoutes(app)

	log.Info().Msg("Server start on :3000 port")
	err := app.Listen(":3000")
	if err != nil {
		log.Error().Err(err)
	}
}

func setupRoutes(app *fiber.App) {
	app.Get("/", factPage)
	app.Get("/f/", factPage)
	app.Get("/f/:lng/:short?", factPage)

	app.Get("/sitemap.xml", sitemapFile)

	app.Get("/api/:lng/fact", getRandomFact)
	app.Get("/api/:lng/fact/:short", getFactByID)
	app.Get("/api/sentsitemap", sentSitemap)

	app.Get("/ping", pingPong)

	app.Static("/", "./public")

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

	log.Info().Str("Short", fact.Short)

	return c.Render("template/index", fiber.Map{
		"Lang":  lng,
		"Fact":  fact.Fact,
		"Short": fact.Short,
		"Year":  t.Year(),
	})
}

func sitemapFile(c *fiber.Ctx) error {
	sm := sitemapSpider()
	return c.Send(sm.XMLContent())
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

func sentSitemap(c *fiber.Ctx) error {
	sm := sitemapSpider()
	sm.Finalize().PingSearchEngines()
	return nil
}

func sitemapSpider() *stm.Sitemap {
	sm := stm.NewSitemap(0)
	sm.Create()
	sm.SetDefaultHost("https://factios.com")

	err := filepath.Walk("./db/",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if strings.HasSuffix(path, ".json") {
				path = strings.Replace(path, "db/", "", 1)
				path = strings.Replace(path, ".json", "", 1)
				sm.Add(stm.URL{{"loc", "/f/" + path}})
			}
			return nil
		},
	)
	if err != nil {
		log.Error().Err(err)
	}

	return sm
}
