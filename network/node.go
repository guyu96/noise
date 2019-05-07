package network

import (
	"fmt"

	"github.com/cynthiatong/noise"
	"github.com/cynthiatong/noise/identity"
	"github.com/cynthiatong/noise/log"
	"github.com/cynthiatong/noise/protocol"
	"github.com/cynthiatong/noise/skademlia"
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

	if len(peerAddrs) > 0 {
		bootStrapped := false
		for _, addr := range peerAddrs {
			bootstrapPeer, err := GetPeerAtAddress(node, addr)
			if err != nil {
				continue
			}
			// Block the current goroutine until we finish performing a S/Kademlia handshake with our peer.
			skademlia.WaitUntilAuthenticated(bootstrapPeer)

			peerIDs := skademlia.FindNode(
				node,
				protocol.NodeID(node).(skademlia.ID),
				skademlia.BucketSize(),
				8,
			)

			// Print the 16 closest peers to us we have found via the `FIND_NODE` RPC call.
			log.Info().Msgf("Bootstrapped with peers: %+v\n", peerIDs)

			// Print the peers we currently are routed/connected to.
			// log.Info().Msgf("Peers we are connected to: %+v\n", skademlia.Table(node).GetPeers())

			// only bootstrap with one peer
			bootStrapped = true
			break
		}
		if !bootStrapped {
			panic("failed to bootstrap with peer addresses provided")
		}
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
