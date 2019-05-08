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

func randID(ids []skademlia.ID) skademlia.ID {
	n := rand.Int() % len(ids)
	return ids[n]
}

func main() {
	portFlag := flag.Uint("p", 7000, "")
	flag.Parse()
	rand.Seed(time.Now().Unix())

	if *portFlag == 4000 {
		node = network.InitNetworkNode(ip, *portFlag, bsAddr)

		peers := skademlia.LoadIDs(peerFile)
		target := randID(peers)
		found := skademlia.FindNode(node, target, skademlia.BucketSize(), 8)
		log.Info().Msgf("closest peers to target %s: %+v", target.Address(), skademlia.IDAddresses(found))

	} else {
		peers := []skademlia.ID{}
		for i := 0; i < numPeers; i++ {
			var bsAddr string
			if i == 0 {
				bsAddr = "127.0.0.1:7000"
			} else {
				bsAddr = randID(peers).Address()
			}
			node = network.InitNetworkNode(ip, *portFlag, []string{bsAddr})
			*portFlag++
			peers = append(peers, protocol.NodeID(node).(skademlia.ID))
		}
		skademlia.PersistIDs(peerFile, peers)
		select {}
	}
}
