package merge

import (
	"errors"
	"fmt"

	"github.com/draganm/statemate"
	"github.com/samber/lo"
	"github.com/urfave/cli/v2"
	"golang.org/x/exp/slices"
)

func Command() *cli.Command {
	cfg := struct {
		stateFiles *cli.StringSlice
		outputFile string
	}{
		stateFiles: cli.NewStringSlice(),
	}

	return &cli.Command{
		Name: "merge",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:        "input-files",
				EnvVars:     []string{"INPUT_FILES"},
				Required:    true,
				Destination: cfg.stateFiles,
			},
			&cli.StringFlag{
				Name:        "output-file",
				EnvVars:     []string{"OUTPUT_FILE"},
				Required:    true,
				Destination: &cfg.outputFile,
			},
		},
		Action: func(c *cli.Context) error {

			stateFilesWithError := lo.Map(cfg.stateFiles.Value(), func(stateFile string, _ int) lo.Tuple2[*statemate.StateMate[uint64], error] {
				sm, err := statemate.Open[uint64](stateFile, statemate.Options{})
				if err != nil {
					return lo.Tuple2[*statemate.StateMate[uint64], error]{nil, fmt.Errorf("could not open state file: %w", err)}
				}
				return lo.Tuple2[*statemate.StateMate[uint64], error]{sm, nil}
			})

			err := lo.Reduce(stateFilesWithError, func(err error, sf lo.Tuple2[*statemate.StateMate[uint64], error], _ int) error {
				return errors.Join(err, sf.B)
			}, nil)

			if err != nil {
				return fmt.Errorf("could not open state files: %w", err)
			}

			defer func() {
				lo.ForEach(stateFilesWithError, func(sf lo.Tuple2[*statemate.StateMate[uint64], error], _ int) {
					if sf.A != nil {
						sf.A.Close()
					}
				})
			}()

			stateFiles := lo.Map(stateFilesWithError, func(sf lo.Tuple2[*statemate.StateMate[uint64], error], _ int) *statemate.StateMate[uint64] {
				return sf.A
			})

			slices.SortFunc(stateFiles, func(a, b *statemate.StateMate[uint64]) bool {
				afi, bfi := a.GetFirstIndex(), b.GetFirstIndex()
				return afi < bfi
			})

			ranges := lo.Map(stateFiles, func(sf *statemate.StateMate[uint64], _ int) lo.Tuple2[uint64, uint64] {
				return lo.Tuple2[uint64, uint64]{sf.GetFirstIndex(), sf.GetLastIndex()}
			})

			for i := range stateFiles[1:] {
				if ranges[i].B != (ranges[i+1].A - 1) {
					return fmt.Errorf("files are not adjacent")
				}
			}

			of, err := statemate.Open[uint64](cfg.outputFile, statemate.Options{})
			if err != nil {
				return fmt.Errorf("could not open output file: %w", err)
			}

			for _, sf := range stateFiles {
				for i := sf.GetFirstIndex(); i <= sf.GetLastIndex(); i++ {
					err := sf.Read(i, func(data []byte) error {
						return of.Append(i, data)
					})

					if err != nil {
						return fmt.Errorf("could not write %d: %w", i, err)
					}
				}
			}

			return nil
		},
	}

}
