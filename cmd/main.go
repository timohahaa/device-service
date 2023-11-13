package main

import (
	"fmt"
	"time"

	"github.com/timohahaa/device-service/config"
	"github.com/timohahaa/device-service/internal/filescanner"
)

func main() {
	config, err := config.NewConfig("config/config.yaml")
	if err != nil {
		panic(err)
	}
	fmt.Println(config)

	s := filescanner.NewScanner(time.Second, config.Scanner.InputDirectoryAbsolutePath, config.Scanner.OutputDirectoryAbsolutePath)
	s.Test()
}
