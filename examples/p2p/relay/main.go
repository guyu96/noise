package main

import (
	"bufio"
	"flag"
	"math/rand"
	"os"
	"time"

	"github.com/guyu96/noise"
	"github.com/guyu96/noise/log"
	"github.com/guyu96/noise/network"
	"github.com/guyu96/noise/protocol"
	"github.com/guyu96/noise/relay"
	kad "github.com/guyu96/noise/skademlia"
)

var (
	ip         = "127.0.0.1"
	bsAddr     = []string{"127.0.0.1:8000"}
	numPeers   = 30
	numBsPeers = 16
	peerFile   = "peers.txt"
	node       *noise.Node
	relayCh    chan relay.Message
)

func randID(ids []kad.ID) kad.ID {
	n := rand.Int() % len(ids)
	return ids[n]
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func randBsAddrs(ids []kad.ID) []string {
	n := minInt(len(ids), numBsPeers)
	bsAddrs := make([]string, n)
	for i := 0; i < n; i++ {
		bsAddrs[i] = randID(ids).Address()
	}
	return bsAddrs
}

func main() {
	portFlag := flag.Uint("p", 8000, "")
	flag.Parse()
	rand.Seed(time.Now().Unix())
	reader := bufio.NewReader(os.Stdin)

	if *portFlag == 4000 {
		peers := kad.LoadIDs(peerFile)
		node, relayCh, _ = network.InitNetwork(ip, *portFlag, randBsAddrs(peers), true, false)
		for {
			input, err := reader.ReadString('\n')
			if err != nil {
				panic(err)
			}
			toID := randID(peers)
			msg := relay.NewMessage(protocol.NodeID(node).(kad.ID), toID, []byte(input))
			relay.ToPeer(node, toID, *msg)
		}
	} else {
		peers := []kad.ID{}
		for i := 0; i < numPeers; i++ {
			node, relayCh, _ = network.InitNetwork(ip, *portFlag, randBsAddrs(peers), true, false)
			go func() {
				for {
					select {
					case msg := <-relayCh:
						log.Info().Msgf("new relay msg: %s", msg.Data)
					}
				}
			}()
			*portFlag++
			peers = append(peers, protocol.NodeID(node).(kad.ID))
		}
		kad.PersistIDs(peerFile, peers)
		select {}
	}
}
