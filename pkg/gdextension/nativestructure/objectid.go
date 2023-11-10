package nativestructure

import (
	"sync/atomic"
)

func (o ObjectID) compare(other ObjectID) int {
	if o.Id < other.Id {
		return -1
	} else if o.Id == other.Id {
		return 0
	}
	return 1
}

func (o ObjectID) IsRefCounted() bool {
	return (uint64(o.Id) & (uint64(1) << 63)) != 0
}

func (o ObjectID) IsValid() bool {
	return uint64(o.Id) != uint64(0)
}

func (o ObjectID) IsNull() bool {
	return uint64(o.Id) == uint64(0)
}

var (
	lastObjectIDValue uint64
)

func NewObjectID() ObjectID {
	v := atomic.AddUint64(&lastObjectIDValue, 1)

	return ObjectID{Id: v}
}
