package skademlia

import (
	"bytes"
	"fmt"
	"math/bits"

	"github.com/cynthiatong/noise"
	"github.com/cynthiatong/noise/payload"
	"github.com/cynthiatong/noise/protocol"
	"github.com/pkg/errors"
	"golang.org/x/crypto/blake2b"
)

var (
	_ protocol.ID = (*ID)(nil)
)

// ID contains S/Kademlia node identity & address info.
type ID struct {
	address   string
	publicKey []byte

	hash  []byte // publicKey hash
	nonce []byte
}

func NewID(address string, publicKey, nonce []byte) ID {
	hash := blake2b.Sum256(publicKey)

	return ID{
		address:   address,
		publicKey: publicKey,

		hash:  hash[:],
		nonce: nonce,
	}
}

func (a ID) Read(reader payload.Reader) (msg noise.Message, err error) {
	a.address, err = reader.ReadString()
	if err != nil {
		return nil, errors.Wrap(err, "skademlia: failed to deserialize ID address")
	}

	a.publicKey, err = reader.ReadBytes()
	if err != nil {
		return nil, errors.Wrap(err, "skademlia: failed to deserialize ID public key")
	}

	hash := blake2b.Sum256(a.publicKey)
	a.hash = hash[:]

	a.nonce, err = reader.ReadBytes()
	if err != nil {
		return nil, errors.Wrap(err, "skademlia: failed to deserialize ID nonce")
	}

	return a, nil
}

func (a ID) Write() []byte {
	return payload.NewWriter(nil).
		WriteString(a.address).
		WriteBytes(a.publicKey).
		WriteBytes(a.nonce).
		Bytes()
}

func (a ID) PublicKey() []byte {
	return a.publicKey
}

func (a ID) Hash() []byte {
	return a.hash
}

func (a ID) Address() string {
	return a.address
}

func (a ID) Equals(other protocol.ID) bool {
	if other, ok := other.(ID); ok {
		return bytes.Equal(a.hash, other.hash)
	}

	return false
}

func (a ID) String() string {
	return fmt.Sprintf("S/Kademlia(address: %s, publicKey: %x, hash: %x, nonce: %x)", a.address, a.publicKey[:16], a.hash[:16], a.nonce)
}

func prefixLen(hash []byte) int {
	for i, b := range hash {
		if b != 0 {
			return i*8 + bits.LeadingZeros8(uint8(b))
		}
	}

	return len(hash)*8 - 1
}

func xor(a, b []byte) []byte {
	if len(a) != len(b) {
		panic("skademlia: len(a) and len(b) must be equal for xor(a, b)")
	}

	c := make([]byte, len(a))

	for i := 0; i < len(a); i++ {
		c[i] = a[i] ^ b[i]
	}

	return c
}

func prefixDiff(a, b []byte, n int) int {
	hash, total := xor(a, b), 0

	for i, b := range hash {
		if n <= 8*i {
			break
		} else if n > 8*i && n < 8*(i+1) {
			shift := 8 - uint(n%8)
			b = b >> shift
		}
		total += bits.OnesCount8(uint8(b))
	}
	return total
}
