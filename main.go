package main

import (
	"fmt"

	config "github.com/thehardcanadian/blogaggregator/internal"
)

func main() {
	cfg, err := config.Read()

	if err != nil {
		fmt.Printf("Reading config: %v\n", err)
	}

	if err := cfg.SetUser("geoff"); err != nil {
		fmt.Println("saving config:", err)
		return
	}

	cfg, err = config.Read()
	if err != nil {
		fmt.Printf("Reading config: %v\n", err)
	}

	fmt.Printf("Config: %v", cfg)

}
