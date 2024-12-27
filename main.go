package main

import (
	"fmt"
	"os"

	"github.com/johndosdos/blog_aggregator/internal/commands"
	"github.com/johndosdos/blog_aggregator/internal/config"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("missing argument. e.g. <command> [arguments]")
		os.Exit(1)
	}

	// Config.Read(Src string)
	newConfig, err := config.Read(".gatorconfig.json")
	if err != nil {
		fmt.Println("%w", err)
		os.Exit(1)
	}

	args := os.Args
	cmd := commands.Command{
		// when using os.Args, we need to start at index 1 because index 0 is the program name
		// index 1 = command name, index 2 = username
		Name: args[1],
		Args: args[2:],
	}

	/* 	// TEST
	   	cmd := commands.Command{
	   		// when using os.Args, we need to start at index 1 because index 0 is the program name
	   		// index 1 = command name, index 2 = username
	   		Name: "login",
	   		Args: []string{"jane"},
	   	} */

	state := &commands.State{Config: &newConfig}

	handlerMap := make(map[string]func(*commands.State, commands.Command) error)
	cmds := commands.Commands{Handlers: handlerMap}
	cmds.Register(cmd.Name, commands.HandlerLogin)

	if err := cmds.Run(state, cmd); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	/* 	// Config.SetUser(Dest, Src string)
	   	err = newConfig.SetUser(".gatorconfig.json", "John")
	   	if err != nil {
	   		fmt.Errorf("%w", err)
	   	}

	   	newConfig, err = config.Read(".gatorconfig.json")
	   	if err != nil {
	   		fmt.Errorf("%w", err)
	   	} */

}
