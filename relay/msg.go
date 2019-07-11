package relay

import (
	"bytes"

	"github.com/guyu96/noise"
	"github.com/guyu96/noise/payload"
	kad "github.com/guyu96/noise/skademlia"
	"github.com/pkg/errors"
	"golang.org/x/crypto/blake2b"
)

const (
	hashSize = blake2b.Size256
)

// Message is a relay message with data.
type Message struct {
	From      kad.ID
	To        kad.ID
	Hash      [hashSize]byte
	Data      []byte
	SeenPeers []byte
}

func (msg Message) Write() []byte {
	writer := payload.NewWriter(nil)
	writer.Write(msg.From.Write())
	writer.Write(msg.To.Write())
	writer.Write(msg.Hash[:])
	// Need to specify number of bytes for reader.ReadBytes method to work
	writer.WriteUint32(uint32(len(msg.Data)))
	writer.Write(msg.Data)
	writer.WriteUint32(uint32(len(msg.SeenPeers)))
	writer.Write(msg.SeenPeers)
	return writer.Bytes()
}

func (msg Message) Read(reader payload.Reader) (noise.Message, error) {
	From, err := kad.ID{}.Read(reader)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode From ID")
	}
	msg.From = From.(kad.ID)

	To, err := kad.ID{}.Read(reader)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode To ID")
	}
	msg.To = To.(kad.ID)

	var hash [hashSize]byte
	for i := 0; i < hashSize; i++ {
		h, err := reader.ReadByte()
		if err != nil {
			return nil, err
		}
		hash[i] = h
	}
	msg.Hash = hash

	data, err := reader.ReadBytes()
	if err != nil {
		return nil, err
	}
	msg.Data = data

	seenPeers, err := reader.ReadBytes()
	if err != nil {
		return nil, err
	}
	msg.SeenPeers = seenPeers

	return msg, nil
}

func (msg *Message) isSeenByPeer(peerID kad.ID) bool {
	if msg.From.Equals(peerID) {
		return true
	}
	hashSize := hashSize
	idHash := peerID.Hash()
	numSeenPeers := len(msg.SeenPeers) / hashSize
	for i := 0; i < numSeenPeers; i++ {
		peerIDHash := msg.SeenPeers[i*hashSize : (i+1)*hashSize]
		if bytes.Equal(idHash, peerIDHash) {
			return true
		}
	}
	return false
}

func (msg *Message) addSeenPeer(peerID kad.ID) {
	msg.SeenPeers = append(msg.SeenPeers, peerID.Hash()...)
}

func (msg *Message) generateHash() {
	bytes := append(msg.From.Hash(), msg.To.Hash()...)
	bytes = append(bytes, msg.Data...)
	msg.Hash = blake2b.Sum256(bytes)
}

// NewMessage creates a new Message instance.
func NewMessage(from kad.ID, to kad.ID, data []byte) Message {
	msg := Message{
		From: from,
		To:   to,
		Data: data,
	}
	msg.generateHash()
	return msg
}
