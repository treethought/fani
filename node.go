package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"log"
	mrand "math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/multiformats/go-multiaddr"

	logging "github.com/ipfs/go-log/v2"
)

const rocProtocol = "/roc/0.1.0"

func (n Node) handleStream(stream network.Stream) {
	fmt.Println("Got a new stream!", stream.ID(), stream.Protocol())
	fmt.Println("direction: ", stream.Stat().Direction)

	// Create a buffer stream for non blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	go readData(rw)
	// go writeData(rw)

	// 'stream' will stay open until you close it (or the other side closes it).
}

func readData(rw *bufio.ReadWriter) {
	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from buffer", err)
			panic(err)
		}

		if str == "" {
			return
		}
		if str != "\n" {
			// Green console colour: 	\x1b[32m
			// Reset console colour: 	\x1b[0m
			fmt.Printf("\x1b[32m%s\x1b[0m> ", str)
		}

	}
}

func writeData(rw *bufio.ReadWriter) {
	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from stdin")
			panic(err)
		}

		_, err = rw.WriteString(fmt.Sprintf("%s\n", sendData))
		if err != nil {
			fmt.Println("Error writing to buffer")
			panic(err)
		}
		err = rw.Flush()
		if err != nil {
			fmt.Println("Error flushing buffer")
			panic(err)
		}
	}
}

func (n Node) startReadLoop() {
	stdReader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from stdin")
			panic(err)
		}

		for _, p := range n.h.Peerstore().Peers() {
			if p == n.h.ID() {
				continue
			}
			stream, err := n.h.NewStream(context.TODO(), p, protocol.ID(rocProtocol))
			if err != nil {
				fmt.Println("failedto create stream to: ", p.String())
			}
			rw := bufio.NewWriter(stream)
			_, err = rw.WriteString(fmt.Sprintf("%s\n", sendData))
			if err != nil {
				fmt.Println("Error writing to buffer")
				panic(err)
			}
			err = rw.Flush()
			if err != nil {
				fmt.Println("Error flushing buffer")
				panic(err)
			}
		}

	}
}

type Node struct {
	h      host.Host
	log    logging.EventLogger
	client bool
}

func NewNode(port int, randomness io.Reader, client bool) Node {
	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, randomness)
	if err != nil {
		log.Fatal(err)
	}

	addr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("addr: ", addr)

	host, err := libp2p.New(
		libp2p.ListenAddrs(addr),
		libp2p.Identity(prvKey),
	)
	if err != nil {
		log.Fatal(err)
	}

	node := Node{h: host, log: logging.Logger("node"), client: client}

	peerInfo := peer.AddrInfo{
		ID:    host.ID(),
		Addrs: host.Addrs(),
	}

	fmt.Println("node ID", host.ID())
	node.log.Info("listen addresses:", host.Addrs())

	addrs, err := peer.AddrInfoToP2pAddrs(&peerInfo)
	if err != nil {
		node.log.Fatal(err)
	}
	fmt.Println("libp2p node address", addrs[0])

	return node
}

func (n Node) HandlePeerFound(pi peer.AddrInfo) {
	if pi.ID == n.h.ID() {
		return
	}
	fmt.Println("discovered new peer:", pi.ID.Pretty())
	n.h.Peerstore().AddAddr(pi.ID, pi.Addrs[0], time.Hour*1)

	err := n.h.Connect(context.Background(), pi)
	if err != nil {
		fmt.Printf("error connecting to peer %s: %s\n", pi.ID.Pretty(), err)
	}
}

func (n Node) startMdns() error {
	fmt.Println("starting mdns")
	mdns := mdns.NewMdnsService(n.h, "roc", n)
	return mdns.Start()
}

func (n Node) Start(ctx context.Context, handler network.StreamHandler) {

	err := n.startMdns()
	if err != nil {
		fmt.Println("failed to start mdns: ", err)
	}

	n.h.SetStreamHandler(rocProtocol, n.handleStream)
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

	fmt.Println("waiting for incoming conenction")

	go n.startReadLoop()

}

func (n Node) Stop() {
	if err := n.h.Close(); err != nil {
		log.Fatal(err)
	}
}

// doEcho reads a line of data a stream and writes it back
func doEcho(s network.Stream) error {
	buf := bufio.NewReader(s)
	str, err := buf.ReadString('\n')
	if err != nil {
		return err
	}

	fmt.Printf("read: %s", str)
	_, err = s.Write([]byte(str))
	return err
}

func start() {

	logging.SetAllLoggers(logging.LevelInfo)
	logging.SetLogLevel("rendezvous", "info")

	client := flag.Bool("client", false, "--client")

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

	node := NewNode(0, r, *client)
	node.Start(context.Background(), func(s network.Stream) {
		fmt.Println("handling stream")
		doEcho(s)
		// err := doEcho(s)
		// if err != nil {
		// 	node.log.Error(err)
		// 	s.Reset()
		// } else {
		// 	s.Close()
		// }
	})

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	node.Stop()
}
