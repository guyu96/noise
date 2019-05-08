package network

import (
	"fmt"
	"sync"

	"github.com/cynthiatong/noise"
	"github.com/cynthiatong/noise/log"
	"github.com/cynthiatong/noise/payload"
	"github.com/cynthiatong/noise/protocol"
	kad "github.com/cynthiatong/noise/skademlia"
	"github.com/pkg/errors"
)

const (
	RelayMsgQueueLength = 100
)

type Relay struct {
	relayOpcode   noise.Opcode
	RelayMsgQueue chan RelayMessage

	relayMsgSeen *sync.Map
}

func NewRelay() *Relay {
	r := &Relay{
		relayOpcode:   noise.RegisterMessage(noise.NextAvailableOpcode(), (*RelayMessage)(nil)),
		RelayMsgQueue: make(chan RelayMessage, RelayMsgQueueLength),
		relayMsgSeen:  new(sync.Map),
	}

	return r
}

// relayMesageThroughPeer relays a custom RelayMessage through peer with the given ID.
func (r *Relay) relayThroughPeer(node *noise.Node, peerID kad.ID, msg RelayMessage) error {
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

	// Send lookup request.
	err = peer.SendMessage(msg)
	if err != nil {
		return err
	}
	return nil
}

// SYNC (might take some time)
func (r *Relay) relayToPeer(node *noise.Node, targetID kad.ID, msg RelayMessage) {
	foundIDs := kad.FindNode(node, targetID, kad.BucketSize(), 8)
	for _, id := range foundIDs {
		if err := r.relayThroughPeer(node, id, msg); err != nil {
			log.Error().Msgf("relay failed %s", err)
		} else if id.Equals(msg.target) {
			break
		}
	}
}

func (r *Relay) handleRelayMessage(node *noise.Node, peer *noise.Peer) {
	for {
		select {
		case msg := <-peer.Receive(r.relayOpcode):
			rl := msg.(RelayMessage)

			if _, seen := r.relayMsgSeen.LoadOrStore(string(rl.hash), struct{}{}); !seen {
				if rl.target.Equals(protocol.NodeID(node)) {
					r.RelayMsgQueue <- rl
				} else {
					closestIDs := kad.FindClosestPeers(kad.Table(node), rl.target.Hash(), 8)
					for _, id := range closestIDs {
						r.relayThroughPeer(node, id.(kad.ID), rl)
						if id.Equals(rl.target) {
							break
						}
					}
				}
			}
		default:
			log.Warn().Msgf("gossip channel full")
		}
	}
}

// RelayMessage is a message to be relayed to target node with custom payload.
type RelayMessage struct {
	source  kad.ID
	target  kad.ID
	payload []byte
	hash    []byte
}

func (r RelayMessage) Write() []byte {
	writer := payload.NewWriter(nil)
	writer.Write(r.source.Write())
	writer.Write(r.target.Write())
	writer.Write(r.payload)
	writer.Write(r.hash)
	return writer.Bytes()
}

func (r RelayMessage) Read(reader payload.Reader) (noise.Message, error) {
	source, err := kad.ID{}.Read(reader)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode source ID")
	}
	r.source = source.(kad.ID)

	target, err := kad.ID{}.Read(reader)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode target ID")
	}
	r.target = target.(kad.ID)

	data, err := reader.ReadBytes()
	if err != nil {
		return nil, err
	}
	r.payload = data

	data, err = reader.ReadBytes()
	if err != nil {
		return nil, err
	}
	r.hash = data

	return r, err
}
