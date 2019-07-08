package relay

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/guyu96/noise"
	"github.com/guyu96/noise/log"
	"github.com/guyu96/noise/protocol"
	kad "github.com/guyu96/noise/skademlia"
)

const (
	relayChanSize = 1024 // default relay message buffer size
	spreadFactor  = 2    // default relay spread factor
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
		RelayChan:    make(chan Message, relayChanSize),
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
			relayMsg := msg.(Message)

			if relayMsg.To.Equals(protocol.NodeID(node)) {
				b.RelayChan <- relayMsg
			} else {
				b.relayMsgMutex.Lock()
				if _, seen := b.relayMsgSeen[string(relayMsg.Hash[:])]; !seen {
					// Set seen flag to prevent relay loop, even though relay message also contains information on seen peers.
					b.relayMsgSeen[string(relayMsg.Hash[:])] = struct{}{}
					ToPeer(node, relayMsg)
					log.Info().Msgf("%v relaying msg", node.InternalPort())
				}
				b.relayMsgMutex.Unlock()
			}
		}
	}
}

func (b *block) GetRelayChan() chan Message {
	return b.RelayChan
}

// relayThroughPeer relays a message through peer with the given ID.
func relayThroughPeer(node *noise.Node, peerID kad.ID, msg Message, errChan chan error) {
	log.Info().Msgf("%v relay through %v", node.InternalPort(), peerID.Address())
	if peerID.Equals(protocol.NodeID(node)) {
		errChan <- errors.New("cannot relay msg to ourselves")
		return
	}

	peer := protocol.Peer(node, peerID)

	if peer == nil {
		peer, err := node.Dial(peerID.Address())
		if err != nil {
			errChan <- errors.New("cannot reach peer")
			return
		}
		kad.WaitUntilAuthenticated(peer)
	}

	errChan <- peer.SendMessage(msg)
}

// ToPeer relays a message to peer synchronously.
func ToPeer(node *noise.Node, msg Message) error {
	if msg.To.Equals(protocol.NodeID(node)) {
		return errors.New("cannot relay msg to ourselves")
	}
	closestIDs := kad.FindClosestPeers(kad.Table(node), msg.To.Hash(), spreadFactor)
	errChan := make(chan error)

	// Determine if direct messaging is possible and filter out seen peers
	toIndex := -1
	unseenIDIndices := []int{}
	for i, id := range closestIDs {
		if id.Equals(msg.To) {
			toIndex = i
			break
		}
		if !msg.isSeenByPeer(id.(kad.ID)) {
			unseenIDIndices = append(unseenIDIndices, i)
		}
	}

	if toIndex != -1 { // send direct message
		log.Info().Msgf("%v sending direct message", node.InternalPort())
		go relayThroughPeer(node, closestIDs[toIndex].(kad.ID), msg, errChan)
		select {
		case err := <-errChan:
			return err
		}
	} else { // relay through peers
		log.Info().Msgf("%v sending indirect relay", node.InternalPort())
		if len(unseenIDIndices) == 0 { // no more peers to relay through
			return errors.New("ran out of peers to relay through")
		}
		for i := range unseenIDIndices {
			go relayThroughPeer(node, closestIDs[i].(kad.ID), msg, errChan)
		}

		numUnseenPeers := uint32(len(unseenIDIndices))
		errCount := uint32(0)
		success := false
		var err error
		// We only require one success and the relay fails only if all peers fail
		for !success && atomic.LoadUint32(&errCount) < numUnseenPeers {
			select {
			case err = <-errChan:
				if err != nil {
					atomic.AddUint32(&errCount, 1)
				} else {
					success = true
					return nil
				}
			}
		}
		return err
	}
}
