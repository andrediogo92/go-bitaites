package timeline

import (
	"bytes"
	"crypto/sha256"
	"net/url"

	"github.com/gogo/protobuf/proto"
)

type Post struct {
	post *PostI
	url *url.URL
}

func (post *Post) PostKey() []byte {
	bbf := new(bytes.Buffer)
	bbf.WriteString(post.post.String() + post.url.EscapedPath())
	return sha256.Sum256(bbf.Bytes())[:]
}

func (post *Post) AsBinary() (ret []byte, err error) {
	ret, err = proto.Marshal(post.post)
	return
}

func PostFromBinary(post []byte) (new *Post, err error) {
	val := &PostI{}
	err = proto.Unmarshal(post, val)
	if err == nil {
		url, err := url.ParseRequestURI(val.Url)
		if err == nil {
			new = &Post{post: val, url: url}
		}
	}
	return
}