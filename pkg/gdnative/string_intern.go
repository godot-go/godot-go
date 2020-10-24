package gdnative

import "sync"

type stringHash uint64

var (
	goStringMap    map[string]String
	godotStringMap map[stringHash]string
	mut            sync.RWMutex
)

func init() {
	goStringMap = map[string]String{}
	godotStringMap = map[stringHash]string{}
}

func internWithGoString(str string) String {
	mut.RLock()
	if gdstr, ok := goStringMap[str]; ok {
		mut.RUnlock()
		return gdstr
	}
	mut.RUnlock()

	mut.Lock()
	gdstr := NewStringFromGoString(str)
	hash := stringHash(gdstr.Hash64())
	goStringMap[str] = gdstr
	godotStringMap[hash] = str
	mut.Unlock()

	return gdstr
}

func internWithGodotString(gdstr String) string {
	hash := stringHash(gdstr.Hash64())
	mut.RLock()
	if str, ok := godotStringMap[hash]; ok {
		mut.RUnlock()
		return str
	}
	mut.RUnlock()

	mut.Lock()
	copy := NewStringCopy(gdstr)
	str := copy.AsGoString()
	goStringMap[str] = copy
	godotStringMap[hash] = str
	mut.Unlock()

	return str
}

func internWithGodotStringName(gdstrname StringName) string {
	return internWithGodotString(gdstrname.GetName())
}

func internClear() {
	mut.Lock()
	godotStringMap = make(map[stringHash]string)
	for k, v := range goStringMap {
		v.Destroy()
		delete(goStringMap, k)
	}
	mut.Unlock()
}
