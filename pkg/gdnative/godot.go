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

//RegisterInitCallback registers funcions to be called after NativeScript initializes.
func RegisterInitCallback(callbacks ...ExtNativescriptInitCallback) {
	initNativescriptCallbacks = append(initNativescriptCallbacks, callbacks...)
}

//RegisterTerminateCallback registers funcions to be called before NativeScript terminates.
func RegisterTerminateCallback(callbacks ...ExtNativescriptTerminateCallback) {
	terminateCallbacks = append(terminateCallbacks, callbacks...)
}
