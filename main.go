package main

import (
	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/jviguy/vespa-stats/api"
	"github.com/jviguy/vespa-stats/db"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	db.Connect()

	app := fiber.New(fiber.Config{
		BodyLimit: 1000 * 1024 * 1024, // this is the default limit of 100MB
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000",
		AllowMethods:     "GET,DELETE,PATCH,POST,PUT",
		AllowHeaders:     "Origin,Authorization,Access-Control-Allow-Origin,Content-Type,Accept,Content-Length,Accept-Language,Accept-Encoding,Connection,",
		AllowCredentials: true,
		MaxAge:           3600,
	}))

	app.Use(logger.New())

	app.Use(jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte("08c14aa51360ba7174ccee0d17800f476a39ef3020a8e9be24f100ac9f46f25e")},
	}))

	app.Group("/api").
		Post("/login", api.Login).
		Post("/process", api.Process)

	err := app.Listen(":8000")
	if err != nil {
		panic(err)
	}
}
