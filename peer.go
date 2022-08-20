package fani

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	ipfslite "github.com/hsanjuan/ipfs-lite"
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/multiformats/go-multiaddr"
	"github.com/suborbital/sat/sat"
)

type FanPeer struct {
	ctx  context.Context
	ipfs *ipfslite.Peer
}

func initIpfs(ctx context.Context) *ipfslite.Peer {
	ds := ipfslite.NewInMemoryDatastore()

	prvKey, _, err := crypto.GenerateKeyPair(crypto.RSA, 2048)
	if err != nil {
		log.Fatal(err)
	}

	listenAddr, err := multiaddr.NewMultiaddr("/ip4/0.0.0.0/tcp/0")
	if err != nil {
		log.Fatal(err)
	}

	h, dht, err := ipfslite.SetupLibp2p(ctx, prvKey, nil, []multiaddr.Multiaddr{listenAddr}, ds, ipfslite.Libp2pOptionsExtra...)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("created libp2p host: ", h.ID().String())
	fmt.Println("listenting on: ", h.Addrs()[0].String())

	lite, err := ipfslite.New(ctx, ds, h, dht, nil)
	if err != nil {
		log.Fatal(err)
	}
	return lite
}

func NewFanPeer() FanPeer {
	ctx := context.TODO()
	lite := initIpfs(ctx)
	return FanPeer{ipfs: lite}
}

func (e FanPeer) Bootstrap() {
	e.ipfs.Bootstrap(ipfslite.DefaultBootstrapPeers())
}

func (e FanPeer) resolveABI(c cid.Cid) (FnABI, error) {

	fmt.Println("resolving ABI: ", c.String())
	r, err := e.ipfs.GetFile(context.TODO(), c)
	if err != nil {
		return FnABI{}, err
	}
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return FnABI{}, err
	}

	abi := FnABI{}
	err = json.Unmarshal(data, &abi)
	if err != nil {
		return FnABI{}, err
	}
	fmt.Println("GOT ABI")
	fmt.Printf("%+v\n", abi)
	return abi, nil

}

func (e FanPeer) getByteCode(abi FnABI) (string, error) {
	fmt.Println("getting bytecode")
	r, err := e.ipfs.GetFile(context.TODO(), abi.ByteCode)
	if err != nil {
		return "", err
	}

	target := filepath.Join("./cache", abi.ByteCode.String())
	f, err := os.Create(target)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	_, err = io.Copy(f, r)
	if err != nil {
		log.Fatal(err)
	}
	return target, nil

}

func (p FanPeer) Execute(c cid.Cid) {
	abi, err := p.resolveABI(c)
	if err != nil {
		fmt.Println("failed to get fn ABI CID: ", err)
		log.Fatal(err)
	}
	bpath, err := p.getByteCode(abi)
	if err != nil {
		log.Fatal("failed to get bytecode")
	}

	ssat := createSat(bpath)
	// defer ssat.Shutdown(context.TODO(), syscall.SIGKILL)

	execStat(ssat, abi.Args)
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

func (p FanPeer) Deploy(path string, id string) cid.Cid {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	bcNode, err := p.ipfs.AddFile(context.TODO(), f, &ipfslite.AddParams{})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("added bytecode", bcNode.Cid().String())

	abi := &FnABI{
		ID:       id,
		ByteCode: bcNode.Cid(),
	}

	abiData, err := json.Marshal(abi)
	if err != nil {
		log.Fatal(err)
	}

	abiNode, err := p.ipfs.AddFile(context.TODO(), bytes.NewReader(abiData), &ipfslite.AddParams{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("added abi", abiNode.Cid().String())

	return abiNode.Cid()
}
