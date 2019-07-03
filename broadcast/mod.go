package broadcast

import (
	"fmt"
	"sync"

	"github.com/guyu96/noise"
	"github.com/guyu96/noise/log"
	"github.com/guyu96/noise/protocol"
	kad "github.com/guyu96/noise/skademlia"
)

const (
	// broadcastChanSize is the default relay msg buffer size
	broadcastChanSize = 100
	// MaxBucketID is the skademlia max bucket ID (index)
	MaxBucketID = 255
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
			b.broadcastMutex.Lock()
			if _, seen := b.broadcastSeen[string(bc.Hash[:])]; !seen {
				b.broadcastSeen[string(bc.Hash[:])] = struct{}{}
				b.BroadcastChan <- bc
				minBucketID := int(bc.PrefixLen)                     // min common prefix
				maxBucketID := kad.Table(node).GetNumOfBuckets() - 1 // maximum common prefix: 256-1
				go SendMessage(node, bc.From, bc.Data, minBucketID, maxBucketID)
			}
			b.broadcastMutex.Unlock()
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
func SendMessage(node *noise.Node, from kad.ID, data []byte, minBucket int, maxBucket int) {
	peers, prefixLens := kad.Table(node).GetBroadcastPeers(minBucket, maxBucket)
	// log.Warn().Msgf("peers %v", peers)
	for i, id := range peers {
		msg := NewMessage(from, prefixLens[i], data)
		if err := broadcastToPeer(node, id.(kad.ID), *msg); err != nil {
			log.Error().Msgf("broadcast failed %s", err)
		} else {
			// log.Info().Msgf("broadcasted to %s", id.(kad.ID).Address())
		}
	}
}
