package main

import (
	"bufio"
	"flag"
	"math/rand"
	"os"
	"sync/atomic"
	"time"

	"github.com/cynthiatong/noise/broadcast"
	"github.com/cynthiatong/noise/log"
	"github.com/cynthiatong/noise/network"
	"github.com/cynthiatong/noise/protocol"
	kad "github.com/cynthiatong/noise/skademlia"
)

var (
	ip         = "127.0.0.1"
	bsAddr     = []string{"127.0.0.1:8000"}
	numPeers   = 30
	numBsPeers = 16
	peerFile   = "peers.txt"
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
		node, _, _ := network.InitNetwork(ip, *portFlag, randBsAddrs(peers), true, true)
		for {
			input, err := reader.ReadString('\n')
			if err != nil {
				panic(err)
			}
			broadcast.SendMessage(node, protocol.NodeID(node).(kad.ID), []byte(input), 0, kad.Table(node).GetNumOfBuckets()-1)
		}
	} else {
		peers := []kad.ID{}
		rcvCount := uint32(0)
		for i := 0; i < numPeers; i++ {
			node, _, bCh := network.InitNetwork(ip, *portFlag, randBsAddrs(peers), true, true)
			go func() {
				for {
					select {
					case msg := <-bCh:
						atomic.AddUint32(&rcvCount, 1)
						log.Info().Msgf("msg %s rcv count %d", msg.Data, atomic.LoadUint32(&rcvCount))
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
