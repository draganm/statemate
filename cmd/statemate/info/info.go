package info

import (
	"fmt"

	"github.com/draganm/statemate"
	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	cfg := struct {
		stateFile string
	}{}

	return &cli.Command{
		Name: "info",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "state",
				EnvVars:     []string{"STATE"},
				Required:    true,
				Destination: &cfg.stateFile,
			},
		},
		Action: func(c *cli.Context) error {
			sm, err := statemate.Open[uint64](cfg.stateFile, statemate.Options{})
			if err != nil {
				return fmt.Errorf("could not open state file: %w", err)
			}

			defer sm.Close()

			fmt.Printf("state file: %s\n", cfg.stateFile)
			fmt.Printf("first index: %d\n", sm.GetFirstIndex())
			fmt.Printf("last index: %d\n", sm.GetLastIndex())
			fmt.Printf("count: %d\n", sm.Count())
			return nil
		},
	}

}
