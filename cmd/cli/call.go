package cli

import (
	"fmt"
	"log"

	"github.com/ipfs/go-cid"
	"github.com/spf13/cobra"
	"github.com/treethought/fani"
)

// callCmd represents the exec command
var callCmd = &cobra.Command{
	Use:   "call",
	Short: "Execute a CID locally",
	Long:  "Execute resolves a CID, retrieves all DAGs required for computation, and executes the function locally.",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		p := fani.NewFanPeer()

		c, err := cid.Parse(args[0])
		if err != nil {
			log.Fatalf("%s is not a valid CID", args[0])
		}

		argCids := []cid.Cid{}
		if len(args) > 1 {
			for _, a := range args[1:] {
				argCids = append(argCids, p.Add(a))
			}
		}

		p.Bootstrap()
		p.StartMdns()
		result := p.Call(c, argCids...)
		fmt.Println("result added to network: ", result.String())

		select {}
	},
}

func init() {
	rootCmd.AddCommand(callCmd)
}
