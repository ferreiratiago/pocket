package cli

import (
	"fmt"
	"net/http"

	"github.com/pokt-network/pocket/rpc"
	"github.com/spf13/cobra"
)

func init() {
	accounCmd := NewSystemCommand()
	accounCmd.Flags().StringVar(&pwd, "pwd", "", "passphrase used by the cmd, non empty usage bypass interactive prompt")
	rootCmd.AddCommand(accounCmd)
}

func NewSystemCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "System",
		Short:   "Commands related to health and troubleshooting of the node instance",
		Aliases: []string{"sys"},
		Args:    cobra.ExactArgs(0),
	}

	cmd.AddCommand(systemCommands()...)

	return cmd
}

func systemCommands() []*cobra.Command {
	cmds := []*cobra.Command{
		{
			Use:     "Health",
			Long:    "Performs a simple liveness check on the node RPC endpoint",
			Aliases: []string{"health"},
			RunE: func(cmd *cobra.Command, args []string) error {

				client, err := rpc.NewClientWithResponses(remoteCLIURL)
				if err != nil {
					return nil
				}
				response, err := client.GetV1HealthWithResponse(cmd.Context())
				if err != nil {
					return unableToConnectToRpc(err)
				}
				statusCode := response.StatusCode()
				if statusCode == http.StatusOK {
					fmt.Printf("✅ RPC reporting healthy status for node @ \033[1m%s\033[0m\n\n%s", remoteCLIURL, response.Body)
					return nil
				}

				return rpcResponseCodeUnhealthy(statusCode, response.Body)
			},
		},
		{
			Use:     "Version",
			Long:    "Queries the node RPC to obtain the version of the software currently running",
			Aliases: []string{"version"},
			RunE: func(cmd *cobra.Command, args []string) error {

				client, err := rpc.NewClientWithResponses(remoteCLIURL)
				if err != nil {
					return err
				}
				response, err := client.GetV1VersionWithResponse(cmd.Context())
				if err != nil {
					return unableToConnectToRpc(err)
				}
				statusCode := response.StatusCode()
				if statusCode == http.StatusOK {
					fmt.Printf("Node @ \033[1m%s\033[0m reports that it's running version: \n\033[1m%s\033[0m\n", remoteCLIURL, response.Body)
					return nil
				}

				return rpcResponseCodeUnhealthy(statusCode, response.Body)
			},
		},
	}
	return cmds
}
