package truncate

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
		Name:        "truncate",
		Description: "truncates the data and index file to the minimal size",
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

			return sm.Truncate()
		},
	}

}
