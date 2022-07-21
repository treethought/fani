package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	mrand "math/rand"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	peerstore "github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/multiformats/go-multiaddr"
)

const rocProtocol = "/roc/0.1.0"

type Node struct {
	h host.Host
}

func NewNode(port int, randomness io.Reader) Node {
	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, randomness)
	if err != nil {
		log.Fatal(err)
	}

	addr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))
	if err != nil {
		log.Fatal(err)
	}

	host, err := libp2p.New(
		libp2p.ListenAddrs(addr),
		libp2p.Identity(prvKey),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("listen addresses:", host.Addrs())

	return Node{h: host}
}

func (n Node) startMdns() error {
	fmt.Println("starting mdns")
	mdns := mdns.NewMdnsService(n.h, "roc", &discoveryNotifee{h: n.h})
	return mdns.Start()
}

func (n Node) Start(ctx context.Context, handler network.StreamHandler) {

	err := n.startMdns()
	if err != nil {
		fmt.Println("failed to start mdns: ", err)
	}

	n.h.SetStreamHandler(rocProtocol, handler)
	// get actual tcp port, in case we started with 0 (random available port)
	var port string
	for _, la := range n.h.Network().ListenAddresses() {
		if p, err := la.ValueForProtocol(multiaddr.P_TCP); err == nil {
			port = p
			break
		}
	}

	if port == "" {
		log.Fatal("unable to find node's actual port")
	}

	peerInfo := peerstore.AddrInfo{
		ID:    n.h.ID(),
		Addrs: n.h.Addrs(),
	}

	addrs, err := peerstore.AddrInfoToP2pAddrs(&peerInfo)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("libp2p node address", addrs[0])

	fmt.Println("waiting for incoming conenction")

}

func (n Node) Stop() {
	if err := n.h.Close(); err != nil {
		log.Fatal(err)
	}
}

// discoveryNotifee gets notified when we find a new peer via mDNS discovery
type discoveryNotifee struct {
	h host.Host
}

// HandlePeerFound connects to peers discovered via mDNS. Once they're connected,
// the PubSub system will automatically start interacting with them if they also
// support PubSub.
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	fmt.Println("discovered new peer:", pi.ID.Pretty())
	err := n.h.Connect(context.Background(), pi)
	if err != nil {
		fmt.Printf("error connecting to peer %s: %s\n", pi.ID.Pretty(), err)
	}
}

func main() {
	// If debug is enabled, use a constant random source to generate the peer ID. Only useful for debugging,
	// off by default. Otherwise, it uses rand.Reader.
	var r io.Reader
	// if *debug {
	if false {
		// Use the port number as the randomness source.
		// This will always generate the same host ID on multiple executions, if the same port number is used.
		// Never do this in production code.
		r = mrand.New(mrand.NewSource(int64(0)))
	} else {
		r = rand.Reader
	}

	node := NewNode(0, r)
	node.Start(context.Background(), func(network.Stream) {
		fmt.Println("handling stream")
	})

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	node.Stop()

}
