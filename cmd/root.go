package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "bazel-cache-oci-registry",
	Short: "A Bazel Remote Cache implementation with an OCI registry backend",
	Long: `The bazel-cache-oci-registry tool implements a Bazel Remote Cache implementation
which uses a OCI registry as a backend.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		cmd.Flags().VisitAll(func(flag *pflag.Flag) {
			if flag.Value.String() == "" && os.Getenv(flag.Name) != "" {
				flag.Value.Set(os.Getenv(flag.Name))
			}
		})
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.EnableTraverseRunHooks = true
}
