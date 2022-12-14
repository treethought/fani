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
	"time"

	ipfslite "github.com/hsanjuan/ipfs-lite"
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/multiformats/go-multiaddr"
	"github.com/suborbital/sat/sat"
)

type FanPeer struct {
	ctx  context.Context
	ipfs *ipfslite.Peer
	p2p  host.Host
}

func initIpfs(ctx context.Context) (*ipfslite.Peer, host.Host) {
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
	return lite, h
}

func NewFanPeer() FanPeer {
	ctx := context.TODO()
	lite, h := initIpfs(ctx)
	return FanPeer{ipfs: lite, p2p: h}
}

func (p FanPeer) HandlePeerFound(pi peer.AddrInfo) {
	if pi.ID == p.p2p.ID() {
		return
	}
	fmt.Println("discovered new peer:", pi.ID.Pretty())
	p.p2p.Peerstore().AddAddr(pi.ID, pi.Addrs[0], time.Hour*1)

	err := p.p2p.Connect(context.Background(), pi)
	if err != nil {
		fmt.Printf("error connecting to peer %s: %s\n", pi.ID.Pretty(), err)
	}
}

func (p FanPeer) StartMdns() error {
	fmt.Println("starting mdns")
	msvc := mdns.NewMdnsService(p.p2p, "roc", p)
	return msvc.Start()
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

func (p FanPeer) Add(r io.Reader) cid.Cid {

	n, err := p.ipfs.AddFile(context.TODO(), r, &ipfslite.AddParams{})
	if err != nil {
		log.Fatal(err)
	}
	return n.Cid()
}

func (p FanPeer) getCids(cids ...cid.Cid) [][]byte {
	result := [][]byte{}

	for _, ac := range cids {
		r, err := p.ipfs.GetFile(context.TODO(), ac)
		if err != nil {
			log.Fatal(err)
		}
		content, err := ioutil.ReadAll(r)
		if err != nil {
			log.Fatal(err)
		}
		result = append(result, content)

	}
	return result

}

func (p FanPeer) Call(fcid cid.Cid, args ...cid.Cid) cid.Cid {
	abi, err := p.resolveABI(fcid)
	if err != nil {
		fmt.Println("failed to get fn ABI CID: ", err)
		log.Fatal(err)
	}
	bpath, err := p.getByteCode(abi)
	if err != nil {
		log.Fatal("failed to get bytecode")
	}

	argsContent := p.getCids(args...)
	result := p.execute(bpath, argsContent)
	fmt.Printf("result:\n%s\n", result)

	r := bytes.NewReader(result)
	resultCID := p.Add(r)

	return resultCID
}

func (p FanPeer) execute(bytecodePath string, args [][]byte) []byte {
	ssat := createSat(bytecodePath)
	// defer ssat.Shutdown(context.TODO(), syscall.SIGKILL)

	return execStat(ssat, args)
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

func execStat(s *sat.Sat, args [][]byte) []byte {
	// TODO: currently only support 1 arg
	// to align with suborbital
	input := []byte{}
	if len(args) > 0 {
		input = args[0]
	}
	resp, err := s.Exec(input)
	if err != nil {
		log.Fatal(err)
	}
	return resp.Output

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
