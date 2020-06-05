package gdnative

const Version = "0.1"

type RegisterMethodBindsCallback func()
type RegisterTypeTagCallback func()
type InitClassCallback func()
type ExtNativescriptInitCallback func()
type ExtNativescriptTerminateCallback func()

var (
	registerMethodBindsCallbacks []RegisterMethodBindsCallback
	registerTypeTagCallbacks     []RegisterTypeTagCallback
	initClassCallbacks           []InitClassCallback
	initNativescriptCallbacks    []ExtNativescriptInitCallback
	terminateCallbacks           []ExtNativescriptTerminateCallback
)

func registerMethodBinds(callbacks ...RegisterMethodBindsCallback) {
	registerMethodBindsCallbacks = append(registerMethodBindsCallbacks, callbacks...)
}

func registerTypeTag(callbacks ...RegisterTypeTagCallback) {
	registerTypeTagCallbacks = append(registerTypeTagCallbacks, callbacks...)
}

//RegisterInitCallback is called for each Godot class that needs to be initialzied as well as to initialize custom code.
func RegisterInitCallback(callbacks ...ExtNativescriptInitCallback) {
	initNativescriptCallbacks = append(initNativescriptCallbacks, callbacks...)
}

func RegisterTerminateCallbacks(callbacks ...ExtNativescriptTerminateCallback) {
	terminateCallbacks = append(terminateCallbacks, callbacks...)
}
