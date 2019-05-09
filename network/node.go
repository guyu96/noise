package network

import (
	"fmt"
	"time"

	"github.com/cynthiatong/noise"
	"github.com/cynthiatong/noise/log"
	"github.com/cynthiatong/noise/protocol"
	"github.com/cynthiatong/noise/relay"
	kad "github.com/cynthiatong/noise/skademlia"
)

const (
	bootstrapTimeout = time.Second * 8
)

func DialPeerAtAddress(node *noise.Node, address string) (*noise.Peer, error) {
	peer, err := node.Dial(address)
	if err != nil {
		return nil, err
	}
	return peer, nil
}

func InitNetworkNode(host string, port uint, peerAddrs []string, bRelay bool) (node *noise.Node, relayCh chan relay.Message) {
	var err error
	// new networking node for this poster
	params := noise.DefaultParams()
	params.Keys = kad.RandomKeys()
	params.Host = host
	params.Port = uint16(port)

	node, err = noise.NewNode(params)
	if err != nil {
		panic(err)
	}

	p := protocol.New()

	if bRelay {
		r := relay.New()
		p.Register(r)
		relayCh = r.GetRelayChan()
	}

	p.Register(kad.New())
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
				bootstrapPeer, err := DialPeerAtAddress(node, addr)
				if err != nil {
					continue
				}
				kad.WaitUntilAuthenticated(bootstrapPeer)
				kad.FindNode(
					node,
					protocol.NodeID(node).(kad.ID),
					kad.BucketSize(),
					8,
				)
				// log.Info().Msgf("bs with: %+v", kad.IDAddresses(ids))

				numPeers := len(kad.Table(node).GetPeers())
				if numPeers >= kad.BucketSize() || numPeers >= len(peerAddrs) {
					close(bsCh)
					break
				}
				// Print the peers we currently are routed/connected to.
				// log.Info().Msgf("Peers we are connected to: %+v\n", kad.Table(node).GetPeers())
			}
		} else {
			close(bsCh) // no bootstrap addresses provided (the first node)
		}
	}()

	select {
	case <-bsCh: // bootstrapped
		log.Info().Msgf("%s:%d bootstrapped", host, port)
	case <-timer.C:
		log.Warn().Msgf("%s:%d bootstrap timeout before finding enough peers", host, port)
	}

	node.OnPeerInit(func(node *noise.Node, peer *noise.Peer) error {
		log.Info().Msg("Peer init")

		peer.OnConnError(func(node *noise.Node, peer *noise.Peer, err error) error {
			log.Info().Msgf("Got an error: %v", err)
			return nil
		})

		// 	peer.OnDisconnect(func(node *noise.Node, peer *noise.Peer) error {
		// 		log.Info().Msgf("Peer %s has disconnected.", protocol.PeerID(peer).(kad.ID).Address())
		// 		return nil
		// 	})

		// peer.AfterMessageReceived(func(node *noise.Node, peer *noise.Peer) error {
		// 	log.Info().Msg("msg rcv")
		// 	return nil
		// })

		// peer.BeforeMessageReceived(func(node *noise.Node, peer *noise.Peer, msg []byte) ([]byte, error) {
		// 	log.Info().Msgf("msg bf: %x", msg)
		// 	return msg, nil
		// })

		return nil
	})

	return node, relayCh
}
