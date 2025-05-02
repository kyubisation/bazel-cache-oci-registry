package cmd

import (
	"bazel-cache-oci-registry/cache"
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/credentials"
	"oras.land/oras-go/v2/registry/remote/retry"
)

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
	RunE: serve,
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().StringVar(&repository, "repository", "", "Fully qualified repository to be used as a cache")
	serverCmd.Flags().StringVar(&username, "username", "", "Username for the OCI registry")
	serverCmd.Flags().StringVar(&password, "password", "", "Password for the OCI registry")
	serverCmd.Flags().StringVar(&token, "token", "", "Access token for the OCI registry")
}

func serve(cmd *cobra.Command, args []string) error {
	if len(repository) == 0 {
		return fmt.Errorf("repository must be configured")
	}

	var authCredentials *auth.Credential
	if len(username+password+token) != 0 {
		authCredentials = &auth.Credential{}
		if len(username) != 0 {
			authCredentials.Username = username
		}
		if len(password) != 0 {
			authCredentials.Password = password
		}
		if len(token) != 0 {
			authCredentials.AccessToken = token
		}
	}

	repo, err := remote.NewRepository(repository)
	if err != nil {
		return err
	} else if strings.HasPrefix(repo.Reference.Registry, "127.0.0.1") || strings.HasPrefix(repo.Reference.Registry, "localhost") {
		repo.PlainHTTP = true
	}
	ctx := cmd.Context()
	if authCredentials != nil {
		repo.Client = &auth.Client{
			Client:     retry.DefaultClient,
			Cache:      auth.NewCache(),
			Credential: auth.StaticCredential(repo.Reference.Registry, *authCredentials),
		}

		// Test the authentication
		registry, err := remote.NewRegistry(repo.Reference.Registry)
		if err != nil {
			return err
		}
		ctx = auth.AppendRepositoryScope(ctx, repo.Reference, auth.ActionPull, auth.ActionPush)
		err = credentials.Login(ctx, credentials.NewMemoryStore(), registry, *authCredentials)
		if err != nil {
			return err
		}
	}

	server := &http.Server{
		Addr:    ":8080",
		Handler: cache.CreateHandler(cache.NewOras(ctx, repo)),
	}
	return server.ListenAndServe()
}
