package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// copyCommandFlags copies flags from source command to destination
func copyCommandFlags(dst, src *cobra.Command) {
	src.Flags().VisitAll(func(flag *pflag.Flag) {
		dst.Flags().AddFlag(flag)
	})
}
