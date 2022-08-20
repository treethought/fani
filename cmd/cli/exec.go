package cli

import (
	"log"

	"github.com/ipfs/go-cid"
	"github.com/spf13/cobra"
	"github.com/treethought/fani"
)

// execCmd represents the exec command
var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "Execute a CID locally",
	Long:  "Execute resolves a CID, retrieves all DAGs required for computation, and executes the function locally.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		p := fani.NewFanPeer()

		c, err := cid.Parse(args[0])
		if err != nil {
			log.Fatalf("%s is not a valid CID", args[0])
		}

		p.Bootstrap()
		p.Execute(c)
	},
}

func init() {
	rootCmd.AddCommand(execCmd)
}
