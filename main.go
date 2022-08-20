package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ipfs/go-cid"
)

const sample = "Qmf9PXdLwdkEdqzmDbCtBCGoxCp6vNPXCr1XQ9fMXxYprs"

func addSample(h Host) cid.Cid {
	f, err := os.Open("./helloworld.wasm")
	if err != nil {
		log.Fatal(err)
	}

	wasmRawCID, err := h.sh.Add(f)
	f.Close()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("added bytecode", wasmRawCID)

	wasmCID, err := cid.Parse(wasmRawCID)
	if err != nil {
		log.Fatal(err)
	}

	abi := &FnABI{
		ID:       "helloworld",
		ByteCode: wasmCID,
	}

	abiData, err := json.Marshal(abi)
	if err != nil {
		log.Fatal(err)
	}

	abiRawCID, err := h.sh.DagPut(abiData, "json", "cbor")
	f.Close()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("added abi", abiRawCID)

	abiCID, err := cid.Parse(abiRawCID)
	if err != nil {
		log.Fatal(err)
	}

	return abiCID
}

func main() {
	noSeed := flag.Bool("no-seed", false, "--no-seed")

	h := NewHost()

	if !*noSeed {
		sampleCID := addSample(h)
		fmt.Println("helloworld sample deployed to:", sampleCID.String())
		h.ExecuteCID(sampleCID)

	}

}
