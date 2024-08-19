package api

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jviguy/vespa-stats/db"
	"golang.org/x/crypto/bcrypt"
	"time"
)

func Login(c *fiber.Ctx) error {
	email := c.FormValue("email")
	pass := c.FormValue("password")
	rows, err := db.DB.Query(context.Background(), "SELECT * FROM users WHERE email = $1", email)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[db.User])
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Email or password is incorrect."})
	}
	err = bcrypt.CompareHashAndPassword(user.Password, []byte(pass))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Email or password is incorrect."})
	}
	// Create the Claims
	claims := jwt.MapClaims{
		"email": user.Email,
		"role":  user.Role,
		"exp":   time.Now().Add(time.Hour * 72).Unix(),
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte("08c14aa51360ba7174ccee0d17800f476a39ef3020a8e9be24f100ac9f46f25e"))
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(fiber.Map{"token": t})
}
