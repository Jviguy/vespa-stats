package api

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/gofiber/fiber/v2"
	"github.com/gofrs/uuid/v5"
	"github.com/golang-jwt/jwt/v5"
	pgxuuid "github.com/jackc/pgx-gofrs-uuid"
	"github.com/jviguy/vespa-stats/db"
	"github.com/redraskal/r6-dissect/dissect"
)

func Process(c *fiber.Ctx) error {
	if c.Locals("user") == nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	name := claims["email"].(string)
	webfile, err := c.FormFile("matchReplay[]")
	if err != nil {
		return err
	}
	filename := webfile.Filename
	file, err := webfile.Open()
	if err != nil {
		panic(err)
	}
	dat, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	matchName := filename[:len(filename)-4]
	zipReader, err := zip.NewReader(bytes.NewReader(dat), int64(len(dat)))
	if err != nil {
		panic(err)
	}
	rounds := len(zipReader.File)
	players := make(map[string]*db.PlayerRoundData)
	for i := 1; i <= len(zipReader.File); i++ {
		zipFile, err := zipReader.Open(matchName + "-R" + fmt.Sprintf("%02d", i) + ".rec")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		r, err := dissect.NewReader(zipFile)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		// Use r.ReadPartial() for faster reads with less data (designed to fill in data gaps in the header)
		// dissect.Ok(err) returns true if the error only pertains to EOF (read was successful)
		if err := r.Read(); !dissect.Ok(err) {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		for _, p := range r.PlayerStats() {
			if _, ok := players[p.Username]; !ok {
				deaths := 0
				if p.Died {
					deaths = 1
				}
				kostInitial := 0
				if deaths == 0 || p.Kills > 0 {
					kostInitial = 1
				}
				OneVx := 0
				if p.OneVx > 0 && r.Header.Teams[p.TeamIndex].Won {
					OneVx++
					fmt.Println(r.Header.Teams[p.TeamIndex].Role)
					fmt.Printf(string(r.Header.Teams[p.TeamIndex].WinCondition))
					fmt.Printf("Player: %s, OneVx: %d, Round: %d\n", p.Username, OneVx, i)
				}
				players[p.Username] = &db.PlayerRoundData{
					Name:         p.Username,
					Team:         "NOT IMPLEMENTED",
					Kills:        p.Kills,
					Deaths:       deaths,
					Assists:      p.Assists,
					EntryFrags:   0,
					EntryFragged: 0,
					Headshots:    p.Headshots,
					Objective:    0,
					Rounds:       rounds,
					KostRounds:   kostInitial,
					OneVx:        OneVx,
				}
			} else {
				players[p.Username].Kills += p.Kills
				if p.Kills > 0 {
					players[p.Username].KostRounds++
				}
				if p.Died {
					players[p.Username].Deaths++
				} else if players[p.Username].KostRounds < i {
					players[p.Username].KostRounds++
				}
				players[p.Username].Assists += p.Assists
				players[p.Username].Headshots += p.Headshots
				if p.OneVx > 0 && r.Header.Teams[p.TeamIndex].Won {
					players[p.Username].OneVx++
					fmt.Println(r.Header.Teams[p.TeamIndex].Role)
					fmt.Println(string(r.Header.Teams[p.TeamIndex].WinCondition))
					fmt.Println(r.Header.Teams[p.TeamIndex].Won)
					fmt.Printf("Player: %s, OneVx: %d, Round: %d\n", p.Username, players[p.Username].OneVx, i)
				}
			}
		}
		players[r.OpeningKill().Username].EntryFrags++
		players[r.OpeningDeath().Username].EntryFragged++
		for _, event := range r.MatchFeedback {
			if event.Type == dissect.DefuserPlantComplete {
				players[event.Username].Objective++
				if players[event.Username].KostRounds < i {
					players[event.Username].KostRounds++
				}
			}
			if event.Type == dissect.DefuserDisableComplete {
				players[event.Username].Objective++
				if players[event.Username].KostRounds < i {
					players[event.Username].KostRounds++
				}
			}
		}
		if i == len(zipReader.File) {
			x, _ := uuid.FromString(r.Header.MatchID)
			var winner int
			if r.Header.Teams[0].Score > r.Header.Teams[1].Score {
				winner = 0
			} else if r.Header.Teams[0].Score < r.Header.Teams[1].Score {
				winner = 1
			} else {
				winner = -1
			}
			_, err := db.DB.Exec(context.Background(),
				"INSERT INTO matches (id, date, teamA, teamB, league, winner, scoreA, scoreB, map)"+
					"VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) ON CONFLICT DO NOTHING",
				pgxuuid.UUID(x), r.Header.Timestamp, "NOT IMPLEMENTED TEAMA", "NOT IMPLEMENTED TEAMB",
				"NOT IMPLEMENTED LEAGUE", winner, r.Header.Teams[0].Score, r.Header.Teams[1].Score,
				r.Header.Map.String(),
			)
			if err != nil {
				panic(err)
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
			}
			for _, player := range players {
				_, err := db.DB.Exec(context.Background(),
					"INSERT INTO players (match_id, name, team, kills, deaths, assists, entry_frags, entry_fragged, headshots, objective, rounds, kost_rounds, onevx)"+
						"VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) ON CONFLICT DO NOTHING",
					pgxuuid.UUID(x), player.Name, player.Team,
					player.Kills, player.Deaths, player.Assists, player.EntryFrags, player.EntryFragged,
					player.Headshots, player.Objective, player.Rounds, player.KostRounds, player.OneVx,
				)
				if err != nil {
					panic(err)
				}
			}
		}
	}

	return c.SendString("Welcome " + name)
}
