package main

import (
	"log"

	"github.com/andras-szesztai/social/internal/db"
	"github.com/andras-szesztai/social/internal/env"
	"github.com/andras-szesztai/social/internal/store"
)

func main() {
	addr := env.GetString("DB_ADDR", "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable")
	conn, err := db.NewDB(addr, 25, 25, "15m")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("database connection pool established")
	defer conn.Close()

	store := store.NewStore(conn)
	err = db.Seed(store, conn)
	if err != nil {
		log.Fatal(err)
	}
}
