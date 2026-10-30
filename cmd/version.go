package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// version is stamped at release time via
// -ldflags "-X github.com/GoogleCloudPlatform/terraformer/cmd.version=<tag>"
// (see .github/workflows/release.yaml). Unstamped builds report "dev".
var version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Terraformer",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Terraformer " + version)
	},
}
