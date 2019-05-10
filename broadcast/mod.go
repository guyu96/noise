package broadcast

import (
	"fmt"
	"sync"

	"github.com/cynthiatong/noise"
	"github.com/cynthiatong/noise/log"
	"github.com/cynthiatong/noise/protocol"
	kad "github.com/cynthiatong/noise/skademlia"
)

const (
	// broadcastChanSize is the default relay msg buffer size
	broadcastChanSize = 100
)

var (
	_ protocol.Block = (*block)(nil)
)

// block stores necessary information for a Relay Message type.
type block struct {
	broadcastOpcode noise.Opcode
	BroadcastChan   chan Message
	broadcastSeen   map[string]struct{}
	broadcastMutex  sync.Mutex
}

// New sets up and returns a new Relay block instance.
func New() *block {
	return &block{
		BroadcastChan: make(chan Message, broadcastChanSize),
		broadcastSeen: map[string]struct{}{},
	}
}

func (b *block) OnRegister(p *protocol.Protocol, node *noise.Node) {
	b.broadcastOpcode = noise.RegisterMessage(noise.NextAvailableOpcode(), (*Message)(nil))
}

func (b *block) OnBegin(p *protocol.Protocol, peer *noise.Peer, node *noise.Node) error {
	go b.handleBroadcastMessage(node, peer)
	return nil
}

func (b *block) OnEnd(p *protocol.Protocol, peer *noise.Peer) error {
	return nil
}

func (b *block) handleBroadcastMessage(node *noise.Node, peer *noise.Peer) {
	for {
		select {
		case msg := <-peer.Receive(b.broadcastOpcode):
			bc := msg.(Message)
			b.BroadcastChan <- bc

			// b.broadcastMutex.Lock()
			// if _, seen := b.broadcastSeen[string(rl.Hash[:])]; !seen {
			// 	b.broadcastSeen[string(rl.Hash[:])] = struct{}{}
			// 	if rl.To.Equals(protocol.NodeID(node)) {
			// 		b.BroadcastChan <- rl
			// 	} else {
			// 		closestIDs := kad.FindClosestPeers(kad.Table(node), rl.To.Hash(), 3)
			// 		// log.Warn().Msgf("closest: %s", closestIDs)
			// 		for _, id := range closestIDs {
			// 			go relayThroughPeer(node, id.(kad.ID), rl)
			// 			if id.Equals(rl.To) {
			// 				break
			// 			}
			// 		}
			// 	}
			// }
			// b.broadcastMutex.Unlock()
			// default:
			// 	log.Warn().Msgf("gossip channel full")
		}
	}
}

func (b *block) GetBroadcastChan() chan Message {
	return b.BroadcastChan
}

// broadcastToPeer broadcasts a custom Message through peer with the given ID.
func broadcastToPeer(node *noise.Node, peerID kad.ID, msg Message) error {
	var err error
	if peerID.Equals(protocol.NodeID(node)) {
		return fmt.Errorf("cannot relay msg to ourselves")
	}

	peer := protocol.Peer(node, peerID)

	if peer == nil {
		peer, err = node.Dial(peerID.Address())

		if err != nil {
			return fmt.Errorf("cannot reach peer")
		}

		kad.WaitUntilAuthenticated(peer)
	}

	// Send relay msg.
	err = peer.SendMessage(msg)
	if err != nil {
		return err
	}
	// log.Info().Msgf("sent to %v", peerID.Address())
	return nil
}

// SendMessage starts broadcasting custom data bytes to the network.
func SendMessage(node *noise.Node, data []byte) {
	peers, prefixLens := kad.Table(node).GetBroadcastPeers()
	// log.Warn().Msgf("peers %v", peers)
	for i, id := range peers {
		myid := protocol.NodeID(node).(kad.ID)
		msg := NewMessage(myid, prefixLens[i], data)
		if err := broadcastToPeer(node, id.(kad.ID), *msg); err != nil {
			log.Error().Msgf("broadcast failed %s", err)
		} else {
			log.Info().Msgf("broadcasted to %s", id.(kad.ID).Address())
		}
	}
}
