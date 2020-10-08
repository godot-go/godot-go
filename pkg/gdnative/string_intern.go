package gdnative

import "sync"

var (
	goStringMap    map[string]*String
	godotStringMap map[*String]string
	mut            sync.RWMutex
)

func init() {
	goStringMap = map[string]*String{}
	godotStringMap = map[*String]string{}
}

func internGoString(str string) *String {
	mut.RLock()
	if gStr, ok := goStringMap[str]; ok {
		mut.RUnlock()
		return gStr
	}
	mut.RUnlock()

	mut.Lock()
	x := NewStringFromGoString(str)
	gStr := &x
	goStringMap[str] = gStr
	godotStringMap[gStr] = str
	mut.Unlock()

	return gStr
}

func internGodotString(gStr *String) string {
	mut.RLock()
	if str, ok := godotStringMap[gStr]; ok {
		mut.RUnlock()
		return str
	}
	mut.RUnlock()

	mut.Lock()
	str := gStr.AsGoString()
	goStringMap[str] = gStr
	godotStringMap[gStr] = str
	mut.Unlock()

	return str
}
