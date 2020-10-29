package function

import (
	"fmt"
	"os"

	"github.com/keti-openfx/openfx-cli/cmd/log"

	"github.com/keti-openfx/openfx-cli/api/grpc"
	"github.com/keti-openfx/openfx-cli/config"
	"github.com/morikuni/aec"
	"github.com/spf13/cobra"
)

func init() {
}

var createCmd = &cobra.Command{
	Use:     `create <NAMESPACES_NAME>`,
	Aliases: []string{"ns"},
	Short:   "Create OpenFx Namespace",
	Long: `
	Create OpenFx Namespace and reads from STDIN for handler(user defined function)'s input(bytes)
	`,
	Example: `  openfx-cli function create namespace`,
	PreRunE: preRunCreate,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("please provide a name for the function")
		}

		namespaceName = args[0]

		if err := runCreate(); err != nil {
			return err
		}
		return nil
	},
}

func preRunCreate(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		log.Fatal("Invalid Namespace name. please describe name of namespace correctly\n")
	}
	namespaceName = args[0]

	gateway = config.GetFxGatewayURL(gateway, "")
	return nil
}

func runCreate() error {

	os.Stderr.WriteString(aec.RedF.Apply(fmt.Sprintf("call runCreate")))

	resp, err := grpc.Create(namespaceName, gateway)
	if err != nil {
		return err
	}

	if resp != "" {
		os.Stdout.WriteString(resp)
	}

	return nil
}
