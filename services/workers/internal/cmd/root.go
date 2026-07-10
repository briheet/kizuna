package cmd

import (
	"context"
	"net/http"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/spf13/cobra"
)

func Execute(ctx context.Context) int {

	var profile bool
	rootCmd := &cobra.Command{
		Use:   "kinuza-workers",
		Short: "Main worker service for Kizuna",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {

			if !profile {
				return nil
			}

			f, perr := os.Create("cpu.pprof")
			if perr != nil {
				return perr
			}

			_ = pprof.StartCPUProfile(f)
			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {

			if !profile {
				return nil
			}

			pprof.StopCPUProfile()

			f, perr := os.Create("mem.pprof")
			if perr != nil {
				return perr
			}
			defer f.Close()

			runtime.GC()
			err := pprof.WriteHeapProfile(f)
			return err
		},
	}

	rootCmd.PersistentFlags().BoolVarP(&profile, "profile", "p", false, "record CPU and Mem pprof")

	go func() {
		_ = http.ListenAndServe("localhost:6070", nil)
	}()

	rootCmd.AddCommand(WorkerCmd(ctx))

	if err := rootCmd.Execute(); err != nil {
		return -1
	}

	return 0
}
