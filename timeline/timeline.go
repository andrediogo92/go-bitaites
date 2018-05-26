package timeline

import (
"crypto"
"crypto/ecdsa"
"crypto/elliptic"
"crypto/rand"
"encoding/binary"






"github.com/Seriyin/go-bitaites/db"
"github.com/dgraph-io/badger"

)

type timeline struct{
	feedMap map[string]*Posts
	timeline []*Post
}

type Id struct {
	user string
	crypto.PublicKey
}

type OwnId struct {
	Id
	crypto.PrivateKey
}

type Posts struct{
	id string
	hashes []string
}

var myid *OwnId
var own *Posts

var t = &timeline{
	nil,
	make([]*Post,40),
}

func init()  {
	dbWrapper := db.GetDBWrapper()
	myuser, err := dbWrapper.GetMyUser([]byte("id-user"))
	switch err {
	case nil: myid, err = dbWrapper.GetMyKeyPair([]byte("keypair" + myuser))
		switch err {
		case nil:
			own, err = dbWrapper.GetOwnPosts([]byte("own-posts"))
			switch err {
			case nil:
			case badger.ErrKeyNotFound:
				//Key not found means generate new Posts
				own = &Posts{
					"own-posts",
					make([]string, 10),
				}
			default:
				panic(err)
			}
		case badger.ErrKeyNotFound:
			//Key not found means generate the keypair.
			myid = GenerateNewId(myuser)
		default:
			panic(err)
		}
	case badger.ErrKeyNotFound:
		panic(err)
	default:
		panic(err)
	}
}

/**
	Generate a new Id based on an elliptic P521 curve.
 */
func GenerateNewId(user string) *OwnId {
	pk, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err == nil {
		return &OwnId{
			Id{
				user,
				pk.Public(),
			},
			pk,
		}
	} else {
		return nil
	}
}

func (timeline *timeline) Key() []byte {
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, 0)
	return bs
}


func GetTimeline() (*timeline){
	return t
}

func GetMyId() (*OwnId) {
	return myid
}

func GetPosts() (*Posts) {
	return own
}

func (posts *Posts) Id() (string) {
	return posts.id
}

func (posts *Posts) Hashes() ([]string) {
	return posts.hashes
}


func TimelineKey() (key []byte) {
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], 0)
	return buf[:]
}