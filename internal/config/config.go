package config

import "os"

type Config struct {
	ListenAddr        string
	DatabaseURL       string
	SessionTTLSeconds int
}

func FromEnv() Config {
	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	db := os.Getenv("DATABASE_URL")
	if db == "" {
		db = "file:abquality.db"
	}
	return Config{ListenAddr: addr, DatabaseURL: db, SessionTTLSeconds: 3600}
}
