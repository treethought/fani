package main

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"syscall"

	"github.com/ipfs/go-cid"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/suborbital/sat/sat"
)

type ArgType struct {
	Name  string
	Codec cid.Cid
}

type FnABI struct {
	ID       string
	ByteCode cid.Cid
	Args     []ArgType
}

type Host struct {
	sh   *shell.Shell
	ssat *sat.Sat
}

func NewHost() Host {
	sh := shell.NewShell("http://127.0.0.1:5001")
	sh.BootstrapAddDefault()
	return Host{
		sh:   sh,
		ssat: nil,
	}
}

func createSat(bcPath string) *sat.Sat {
	config, err := sat.ConfigFromRunnableArg(bcPath)
	if err != nil {
		log.Fatal(err)
	}

	s, err := sat.New(config, nil)
	if err != nil {
		log.Fatal(err)
	}

	return s
}

func execStat(s *sat.Sat, args []ArgType) {
	// TODO: args

	resp, err := s.Exec([]byte{})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", resp.Output)

}

func (n Host) ExecuteCID(cid cid.Cid) {
	fmt.Println("executing fn: ", cid.String())

	abi := &FnABI{}
	err := n.sh.DagGet(cid.String(), abi)
	if err != nil {
		fmt.Println("failed to get fn ABI CID: ", err)
	}
	fmt.Printf("%+v\n", abi)

	target := filepath.Join("./cache", abi.ByteCode.String())

	err = n.sh.Get(abi.ByteCode.String(), target)
	if err != nil {
		log.Fatal("failed to get bytecode")
	}

	ssat := createSat(filepath.Join(target))
	defer ssat.Shutdown(context.TODO(), syscall.SIGKILL)

	execStat(ssat, abi.Args)

}
