package main

import (
	"fmt"
	"internal/config"

	"os"

	"github.com/vladimirck/gator/internal/config"
)

func main() {
	cfg, err := config.Read()

	if err != nil {
		fmt.Printf("The configuration file could no be read: %v", err)
		os.Exit(1)
	}

	if cfg.SetUser("vladimir") != nil {
		fmt.Printf("The username coulnot be set: %v", err)
		os.Exit(1)
	}

	cfg, _ = config.Read()

	fmt.Printf("%v", cfg)

}
