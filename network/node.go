package network

import (
	"fmt"
	"time"

	"github.com/cynthiatong/noise"
	"github.com/cynthiatong/noise/identity"
	"github.com/cynthiatong/noise/log"
	"github.com/cynthiatong/noise/protocol"
	"github.com/cynthiatong/noise/skademlia"
)

const (
	bootstrapTimeout = time.Second * 3
)

/** NetworkHandler methods */
func GetNodeKeys(node *noise.Node) identity.Keypair {
	return node.Keys
}

func GetPeerAtAddress(node *noise.Node, address string) (*noise.Peer, error) {
	peer, err := node.Dial(address)
	if err != nil {
		return nil, err
	}
	return peer, nil
}

func InitNetworkNode(host string, port uint, peerAddrs []string) *noise.Node {
	// new networking node for this poster
	params := noise.DefaultParams()
	params.Keys = skademlia.RandomKeys()
	params.Host = host
	params.Port = uint16(port)

	node, err := noise.NewNode(params)
	if err != nil {
		panic(err)
	}

	p := protocol.New()
	p.Register(skademlia.New())
	p.Enforce(node)

	go node.Listen()

	// TODO put filtering methodin util
	i := 0
	ownAddr := fmt.Sprintf("%s:%d", host, port)
	for _, addr := range peerAddrs {
		if addr != ownAddr {
			peerAddrs[i] = addr
			i++
		}
	}
	peerAddrs = peerAddrs[:i]

	timer := time.NewTimer(bootstrapTimeout)
	bsCh := make(chan struct{})

	go func() {
		if len(peerAddrs) > 0 {
			for _, addr := range peerAddrs {
				bootstrapPeer, err := GetPeerAtAddress(node, addr)
				if err != nil {
					continue
				}
				// Block the current goroutine until we finish performing a S/Kademlia handshake with our peer.
				skademlia.WaitUntilAuthenticated(bootstrapPeer)

				skademlia.FindNode(
					node,
					protocol.NodeID(node).(skademlia.ID),
					skademlia.BucketSize(),
					8,
				)
				// log.Info().Msgf("%s:%d bootstrapped with peers", host, port)

				// Print the peers we currently are routed/connected to.
				// log.Info().Msgf("Peers we are connected to: %+v\n", skademlia.Table(node).GetPeers())

				// only bootstrap with one peer
				close(bsCh)
				break
			}
		} else {
			close(bsCh) // no bootstrap addresses provided (the first node)
		}
	}()

	select {
	case <-bsCh: // bootstrapped
		break
	case <-timer.C:
		log.Fatal().Msgf("%s:%d bootstrap timeout", host, port)
	}

	node.OnPeerInit(func(node *noise.Node, peer *noise.Peer) error {
		peer.OnConnError(func(node *noise.Node, peer *noise.Peer, err error) error {
			log.Info().Msgf("Got an error: %v", err)
			return nil
		})

		peer.OnDisconnect(func(node *noise.Node, peer *noise.Peer) error {
			log.Info().Msgf("Peer %s has disconnected.", protocol.PeerID(peer).(skademlia.ID).Address())
			return nil
		})
		return nil
	})

	return node
}
