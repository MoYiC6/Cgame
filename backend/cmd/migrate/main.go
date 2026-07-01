package main

import (
	"fmt"
	"log"
	"os"

	"backend/internal/platform/config"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func main() {
	cfg, err := config.LoadConfig("")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if cfg.DB.DSN == "" {
		log.Fatal("DB_DSN is required")
	}

	db, err := goose.OpenDBWithDriver("pgx", cfg.DB.DSN)
	if err != nil {
		log.Fatalf("goose open: %v", err)
	}
	defer db.Close()

	if len(os.Args) < 2 {
		log.Fatal("usage: go run cmd/migrate/main.go <up|down>")
	}

	cmd := os.Args[1]
	if cmd == "up" {
		if err := goose.Up(db, "migrations"); err != nil {
			log.Fatalf("goose up: %v", err)
		}
	} else if cmd == "down" {
		if err := goose.Down(db, "migrations"); err != nil {
			log.Fatalf("goose down: %v", err)
		}
	} else {
		log.Fatalf("unknown command: %s", cmd)
	}

	fmt.Println("migration completed")
}
