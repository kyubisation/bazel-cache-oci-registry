package cmd

import (
	"bazel-cache-oci-registry/cache"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
	"oras.land/oras-go/v2/registry/remote/auth"
)

var registry string
var repository string
var username string
var password string
var token string

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the cache server",
	Long: `
	
bazel-cache-oci-registry server`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(registry) == 0 {
			return fmt.Errorf("registry must be configured")
		} else if len(repository) == 0 {
			return fmt.Errorf("repository must be configured")
		}

		var credentials *auth.Credential
		if len(username+password+token) != 0 {
			credentials = &auth.Credential{}
			if len(username) != 0 {
				credentials.Username = username
			}
			if len(password) != 0 {
				credentials.Password = password
			}
			if len(token) != 0 {
				credentials.AccessToken = token
			}
		}

		server := &http.Server{
			Addr: ":8080",
			Handler: cache.CreateHandler(
				cache.NewOras(
					cmd.Context(), registry, repository, credentials)),
		}
		return server.ListenAndServe()
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().StringVar(&registry, "registry", "", "Adress of the OCI registry")
	serverCmd.Flags().StringVar(&repository, "repository", "", "Repository name to be used as a cache")
	serverCmd.Flags().StringVar(&username, "username", "", "Username for the OCI registry")
	serverCmd.Flags().StringVar(&password, "password", "", "Password for the OCI registry")
	serverCmd.Flags().StringVar(&token, "token", "", "Access token for the OCI registry")
}
