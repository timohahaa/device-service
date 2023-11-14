package main

import (
	"time"

	"github.com/timohahaa/device-service/config"
	"github.com/timohahaa/device-service/internal/filescanner"
	"github.com/timohahaa/postgres"
)

func main() {
	config, err := config.NewConfig("config/config.yaml")
	if err != nil {
		panic(err)
	}
	//	fmt.Println(config)
	pg, err := postgres.New("postgres://timohahaa:timohahaa1337@localhost:5432/devices", postgres.MaxConnPoolSize(5))
	if err != nil {
		panic(err)
	}

	s := filescanner.NewScanner(pg, time.Second, config.Scanner.InputDirectoryAbsolutePath, config.Scanner.OutputDirectoryAbsolutePath)
	s.Test()
}
