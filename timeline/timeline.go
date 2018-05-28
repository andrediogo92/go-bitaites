package timeline

import (
"crypto"
"crypto/ecdsa"
"crypto/elliptic"
"crypto/rand"
"encoding/binary"






"github.com/Seriyin/go-bitaites/db"
	"github.com/Seriyin/go-bitaites/peer"
	"github.com/dgraph-io/badger"

)

type timeline struct{
	feedMap map[peer.ID]*Posts
	timeline []*Post
}


type OwnId struct {
	peer.ID
	*peer.UserID
	crypto.PrivateKey
}


type Posts struct{
	id peer.ID
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
	myuser, err := dbWrapper.GetMyUser()
	switch err {
	case nil:
		myid, err = dbWrapper.GetMyKeyPair(myuser)
		switch err {
		case nil:
			own, err = dbWrapper.GetOwnPosts()
			switch err {
			case nil:
			case badger.ErrKeyNotFound:
				//Key not found means generate new Posts
				own = &Posts{
					"own-posts",
					make([]string, 0, 10),
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
		uid := peer.NewUserID(user, pk.Public())
		return &OwnId{
			uid.ComputeSha256(),
			uid,
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

func (posts *Posts) Id() (peer.ID) {
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