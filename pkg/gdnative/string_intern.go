package gdnative

import "sync"

type stringHash uint64

var (
	goStringNameMap    map[string]StringName
	godotStringNameMap map[stringHash]string
	mut                sync.RWMutex
)

func init() {
	goStringNameMap = map[string]StringName{}
	godotStringNameMap = map[stringHash]string{}
}

func internWithGoString(str string) String {
	gdsn := internNameWithGoString(str)
	return gdsn.GetName()
}

func internNameWithGoString(str string) StringName {
	mut.RLock()
	if gdsn, ok := goStringNameMap[str]; ok {
		mut.RUnlock()
		return gdsn
	}
	mut.RUnlock()

	mut.Lock()
	gdsn := NewStringNameData(str)
	gds := gdsn.GetName()
	hash := stringHash(gds.Hash64())
	goStringNameMap[str] = gdsn
	godotStringNameMap[hash] = str
	mut.Unlock()

	return gdsn
}

func internWithGodotStringName(gdsn StringName) string {
	gds := gdsn.GetName()
	hash := stringHash(gds.Hash64())
	mut.RLock()
	if str, ok := godotStringNameMap[hash]; ok {
		mut.RUnlock()
		return str
	}
	mut.RUnlock()

	mut.Lock()
	str := gds.AsGoString()
	goStringNameMap[str] = NewStringName(gds)
	godotStringNameMap[hash] = str
	mut.Unlock()

	return str
}

func internWithGodotString(gds String) string {
	gdsn := NewStringName(gds)
	return internWithGodotStringName(gdsn)
}

func internClear() {
	mut.Lock()
	godotStringNameMap = make(map[stringHash]string)
	for k, v := range goStringNameMap {
		v.Destroy()
		delete(goStringNameMap, k)
	}
	mut.Unlock()
}
