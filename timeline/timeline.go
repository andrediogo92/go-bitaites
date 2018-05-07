package timeline

import (
	"encoding/binary"

	"github.com/Seriyin/go-bitaites/db"
	"github.com/golang/protobuf/proto"
)

type Timeline struct{
	timeline *TimelineI
}


func (timeline *Timeline) Key() []byte {
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, timeline.timeline.Id)
	return bs
}

func (timeline *Timeline) AsBinary() (ret []byte, err error) {
	ret, err = proto.Marshal(timeline.timeline)
	return
}


func (timeline *Timeline) FromBinary(encoded []byte) (err error) {
	val := &TimelineI{}
	err = proto.Unmarshal(encoded, val)
	if err == nil {
		timeline.timeline = val
	}
	return
}

func TimelineFromDB(wrapper db.DBWrapper) (*Timeline){
	return &Timeline{&TimelineI{
		uint32(wrapper.GetNewId()),
		map[string]*TimelineI_PostStream{},
		&TimelineI_PostStream{},
		struct{}{},
		make([]byte,0),
		0}}
}
