package timeline

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"net/url"

	"github.com/golang/protobuf/proto"
)

type Post struct {
	post *PostI
	url *url.URL
}

func (post *Post) Key() []byte {
	bs := make([]byte , 8)
	bbf := new(bytes.Buffer)
	binary.LittleEndian.PutUint64(bs, post.post.Stamp)
	bbf.Write(bs)
	bbf.WriteString(post.post.User)
	return sha256.Sum256(bbf.Bytes())[:]
}

func (post *Post) AsBinary() (ret []byte, err error) {
	ret, err = proto.Marshal(post.post)
	return
}

func (post *Post) FromBinary(encoded []byte) (err error) {
	val := &PostI{}
	err = proto.Unmarshal(encoded, val)
	if err == nil {
		rurl, err := url.ParseRequestURI(val.Url)
		if err == nil {
			post.post = val
			post.url = rurl
		}
	}
	return
}