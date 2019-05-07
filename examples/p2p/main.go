package main

import (
	"flag"

	"github.com/cynthiatong/noise"
	"github.com/cynthiatong/noise/network"
	"github.com/cynthiatong/noise/protocol"
	"github.com/cynthiatong/noise/skademlia"
)

var (
	ip     = "127.0.0.1"
	bsAddr = []string{"127.0.0.1:7000"}
	node   *noise.Node
	peers  []skademlia.ID
)

func main() {
	portFlag := flag.Uint("p", 7000, "")
	flag.Parse()

	if *portFlag == 4000 {
		node = network.InitNetworkNode(ip, *portFlag, bsAddr)
		select {}

		// rand.Seed(time.Now().Unix())
		// n := rand.Int() % len(peers)
		// targetID := peers[n]
		// found := skademlia.FindNode(node, targetID, skademlia.BucketSize(), 8)
		// log.Info().Msgf("closest peers: %+v", found)

	} else {
		peers = []skademlia.ID{}
		for i := 0; i < 8; i++ {
			node = network.InitNetworkNode(ip, *portFlag, bsAddr)
			*portFlag++
			peers = append(peers, protocol.NodeID(node).(skademlia.ID))
		}
		select {}
	}
}
