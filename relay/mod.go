package relay

import (
	"fmt"
	"sync"

	"github.com/cynthiatong/noise"
	"github.com/cynthiatong/noise/log"
	"github.com/cynthiatong/noise/protocol"
	kad "github.com/cynthiatong/noise/skademlia"
)

const (
	// RelayChanSize is the default relay msg buffer size
	RelayChanSize = 100
)

var (
	_ protocol.Block = (*block)(nil)
)

// block stores necessary information for a Relay Message type.
type block struct {
	relayOpcode   noise.Opcode
	RelayChan     chan Message
	relayMsgSeen  map[string]struct{}
	relayMsgMutex sync.Mutex
}

// New sets up and returns a new Relay block instance.
func New() *block {
	return &block{
		RelayChan:    make(chan Message, RelayChanSize),
		relayMsgSeen: map[string]struct{}{},
	}
}

func (b *block) OnRegister(p *protocol.Protocol, node *noise.Node) {
	b.relayOpcode = noise.RegisterMessage(noise.NextAvailableOpcode(), (*Message)(nil))
}

func (b *block) OnBegin(p *protocol.Protocol, peer *noise.Peer, node *noise.Node) error {
	go b.handleRelayMessage(node, peer)
	return nil
}

func (b *block) OnEnd(p *protocol.Protocol, peer *noise.Peer) error {
	return nil
}

func (b *block) handleRelayMessage(node *noise.Node, peer *noise.Peer) {
	for {
		select {
		case msg := <-peer.Receive(b.relayOpcode):
			rl := msg.(Message)

			b.relayMsgMutex.Lock()
			if _, seen := b.relayMsgSeen[string(rl.Hash[:])]; !seen {
				b.relayMsgSeen[string(rl.Hash[:])] = struct{}{}
				if rl.To.Equals(protocol.NodeID(node)) {
					b.RelayChan <- rl
				} else {
					closestIDs := kad.FindClosestPeers(kad.Table(node), rl.To.Hash(), 3)
					// log.Warn().Msgf("closest: %s", closestIDs)
					for _, id := range closestIDs {
						go relayThroughPeer(node, id.(kad.ID), rl)
						if id.Equals(rl.To) {
							break
						}
					}
				}
			}
			b.relayMsgMutex.Unlock()
			// default:
			// 	log.Warn().Msgf("gossip channel full")
		}
	}
}

func (b *block) GetRelayChan() chan Message {
	return b.RelayChan
}

// RelayThroughPeer relays a custom Message through peer with the given ID.
func relayThroughPeer(node *noise.Node, peerID kad.ID, msg Message) error {
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

// ToPeer relays a custom Message to peer synchronously.
func ToPeer(node *noise.Node, targetID kad.ID, msg Message) {
	// log.Warn().Msgf("target %s", targetID.Address())
	foundIDs := kad.FindNode(node, targetID, kad.BucketSize(), 3)
	// log.Warn().Msgf("found %s", kad.IDAddresses(foundIDs))
	for _, id := range foundIDs {
		if err := relayThroughPeer(node, id, msg); err != nil {
			log.Error().Msgf("relay failed %s", err)
		} else if id.Equals(msg.To) {
			log.Info().Msgf("directly sent to %s!", id.Address())
			break
		}
	}
}
