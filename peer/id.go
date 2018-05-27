package peer

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/gob"

	"github.com/Seriyin/go-bitaites/peer/keyspace"
)

type ID string

type UserID struct {
	user string
	sk crypto.PublicKey
}

// MatchesPrivateKey tests whether this ID was derived from sk
func (id ID) MatchesPrivateKey(user string, sk ecdsa.PrivateKey) bool {
	return id.MatchesPublicKey(&UserID{user, sk})
}

func (id ID) MatchesPublicKey(userID *UserID) bool {
	return id == userID.ComputeSha256()
}

func (id ID) Equal(other ID) bool {
	return bytes.Equal([]byte(id), []byte(other))
}

func (id ID) Less(other ID) bool {
	a := keyspace.Key{Space: keyspace.XORKeySpace, Bytes: []byte(id)}
	b := keyspace.Key{Space: keyspace.XORKeySpace, Bytes: []byte(other)}
	return a.Less(b)
}

func NewID(user string, sk crypto.PublicKey) ID {
	return (&UserID{user, sk}).ComputeSha256()
}

func NewUserID(user string, sk crypto.PublicKey) *UserID {
	return &UserID{user, sk}
}

func (uid *UserID) ComputeSha256() ID {
	bs := bytes.Buffer{}
	encod := gob.NewEncoder(&bs)
	encod.Encode(uid)
	h := sha256.New()
	h.Write(bs.Bytes())
	return ID(h.Sum(nil))
}

// Closer returns true if a is closer to key than b is
func Closer(a, b ID, key string) bool {
	adist := ID(keyspace.XOR([]byte(a), []byte(key)))
	bdist := ID(keyspace.XOR([]byte(b), []byte(key)))

	return adist.Less(bdist)
}