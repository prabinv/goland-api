package main

import (
	"fmt"
	"github.com/caarlos0/env/v10"
	"log"
	"my-first-api/internal/db"
	"my-first-api/internal/todo"
	"my-first-api/internal/transport"
)

type config struct {
	Port       int    `env:"PORT" envDefault:"8080"`
	DbServer   string `env:"DB_SERVER" envDefault:"localhost"`
	DbPort     int    `env:"DB_PORT" envDefault:"5432"`
	DbUser     string `env:"DB_USER" envDefault:"postgres"`
	DbPassword string `env:"DB_PASSWORD,required,unset"`
	DbName     string `env:"DB_NAME,required" envDefault:"postgres"`
}

func main() {
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
	}
	dbConfig := db.Config{
		Host:     cfg.DbServer,
		Port:     cfg.DbPort,
		Username: cfg.DbUser,
		Password: cfg.DbPassword,
		Database: cfg.DbName,
	}
	dbSvc, err := db.New(dbConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer dbSvc.Close()
	svc := todo.NewTodoService(dbSvc)

	svr := transport.NewServer(svc)

	if err := svr.Serve(); err != nil {
		log.Fatal("error starting http server:", err)
	} else {
		fmt.Println("server listening on :8080")
	}

}
