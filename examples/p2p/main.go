package main

import (
	"flag"
	"math/rand"
	"time"

	"github.com/cynthiatong/noise"
	"github.com/cynthiatong/noise/log"
	"github.com/cynthiatong/noise/network"
	"github.com/cynthiatong/noise/protocol"
	"github.com/cynthiatong/noise/skademlia"
)

var (
	ip       = "127.0.0.1"
	bsAddr   = []string{"127.0.0.1:7000"}
	node     *noise.Node
	numPeers = 20
	peerFile = "peers.txt"
)

func main() {
	portFlag := flag.Uint("p", 7000, "")
	flag.Parse()

	if *portFlag == 4000 {
		node = network.InitNetworkNode(ip, *portFlag, bsAddr)

		peers := skademlia.LoadIDs(peerFile)
		rand.Seed(time.Now().Unix())
		n := rand.Int() % len(peers)
		targetID := peers[n]
		found := skademlia.FindNode(node, targetID, skademlia.BucketSize(), 8)
		log.Info().Msgf("closest peers: %+v", skademlia.IDAddresses(found))

	} else {
		peers := make([]skademlia.ID, numPeers)
		for i := range peers {
			node = network.InitNetworkNode(ip, *portFlag, bsAddr)
			*portFlag++
			peers[i] = protocol.NodeID(node).(skademlia.ID)
		}
		skademlia.PersistIDs(peerFile, peers)
		select {}
	}
}
