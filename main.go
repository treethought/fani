package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	ipfslite "github.com/hsanjuan/ipfs-lite"
	"github.com/ipfs/go-cid"
)

const sample = "Qmf9PXdLwdkEdqzmDbCtBCGoxCp6vNPXCr1XQ9fMXxYprs"

func addSample(ex Executor) cid.Cid {
	f, err := os.Open("./helloworld.wasm")
	if err != nil {
		log.Fatal(err)
	}

	bcNode, err := ex.ipfs.AddFile(context.TODO(), f, &ipfslite.AddParams{})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("added bytecode", bcNode.Cid().String())

	abi := &FnABI{
		ID:       "helloworld",
		ByteCode: bcNode.Cid(),
		Args: []ArgType{
			{Name: "toGreet"},
		},
	}

	abiData, err := json.Marshal(abi)
	if err != nil {
		log.Fatal(err)
	}

	abiNode, err := ex.ipfs.AddFile(context.TODO(), bytes.NewReader(abiData), &ipfslite.AddParams{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("added abi", abiNode.Cid().String())

	return abiNode.Cid()
}

func main() {
	noSeed := flag.Bool("no-seed", false, "--no-seed")

	ex := NewExecutor()

	if !*noSeed {
		sampleCID := addSample(ex)
		fmt.Println("helloworld sample deployed to:", sampleCID.String())
		fmt.Println()

		// argCidRaw, err := h.sh.Add(bytes.NewReader([]byte("jimmy")))
		// f.Close()
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// fmt.Println("added arg", argCidRaw)
		ex.Execute(sampleCID)
	}

}
