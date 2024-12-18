package main

import (
	"fmt"

	"github.com/johndosdos/blog_aggregator/internal/config"
)

func main() {
	// Config.Read(Src string)
	newConfig, err := config.Read(".gatorconfig.json")
	if err != nil {
		fmt.Errorf("%w", err)
	}

	// Config.SetUser(Dest, Src string)
	err = newConfig.SetUser(".gatorconfig.json", "John")
	if err != nil {
		fmt.Errorf("%w", err)
	}

	newConfig, err = config.Read(".gatorconfig.json")
	if err != nil {
		fmt.Errorf("%w", err)
	}

	fmt.Println(newConfig)
}
