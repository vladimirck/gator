package main

import (
	"fmt"

	"github.com/vladimirck/gator/internal/config"

	"os"
)

func main() {
	cfg, err := config.Read()

	if err != nil {
		fmt.Printf("The configuration file could no be read: %v\n", err)
		os.Exit(1)
	}

	if cfg.SetUser("vladimir") != nil {
		fmt.Printf("The username coulnot be set: %v\n", err)
		os.Exit(1)
	}

	cfg, _ = config.Read()

	fmt.Printf("%v", cfg)

}
