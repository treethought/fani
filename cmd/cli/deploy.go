package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/treethought/fani"
)

var wasmPath string
var name string

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy [name] [wasm file]",
	Short: "deploy a function to the network",
	Long: `deploy constructs a functions ABI, consisting of the wasm bytecode, arguments
    and other metadata to the network. The provided path must be a wasm module compatible with suboribtal's sat`,
	Args:       cobra.ExactArgs(2),
	ArgAliases: []string{"name", "wasm"},
	Run: func(cmd *cobra.Command, args []string) {
		name, wasm := args[0], args[1]

		p := fani.NewFanPeer()
		p.Bootstrap()
		p.StartMdns()

		p.Deploy(wasm, name)

		fmt.Println("sitting idle to provide deployed dag")
		fmt.Println("press ctrl-c to quit")
		select {}
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
}
