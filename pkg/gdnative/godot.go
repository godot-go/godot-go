package gdnative

const Version = "0.1-WIP"

type RegisterMethodBindsCallback func()
type RegisterTypeTagCallback func()
type InitClassCallback func()
type ExtNativescriptInitCallback func()
type ExtNativescriptTerminateCallback func()

var (
	registerMethodBindsCallbacks           []RegisterMethodBindsCallback
	registerTypeTagCallbacks               []RegisterTypeTagCallback
	initInternalNativescriptCallbacks      []ExtNativescriptInitCallback
	terminateInternalNativescriptCallbacks []ExtNativescriptTerminateCallback
	initNativescriptCallbacks              []ExtNativescriptInitCallback
	terminateNativescriptCallbacks         []ExtNativescriptTerminateCallback
)

func registerMethodBinds(callbacks ...RegisterMethodBindsCallback) {
	registerMethodBindsCallbacks = append(registerMethodBindsCallbacks, callbacks...)
}

func registerTypeTag(callbacks ...RegisterTypeTagCallback) {
	registerTypeTagCallbacks = append(registerTypeTagCallbacks, callbacks...)
}

func registerInternalInitCallback(callbacks ...ExtNativescriptInitCallback) {
	initInternalNativescriptCallbacks = append(initInternalNativescriptCallbacks, callbacks...)
}

func registerInternalTerminateCallback(callbacks ...ExtNativescriptTerminateCallback) {
	terminateInternalNativescriptCallbacks = append(terminateInternalNativescriptCallbacks, callbacks...)
}

//RegisterInitCallback registers funcions to be called after NativeScript initializes.
func RegisterInitCallback(callbacks ...ExtNativescriptInitCallback) {
	initNativescriptCallbacks = append(initNativescriptCallbacks, callbacks...)
}

//RegisterTerminateCallback registers funcions to be called before NativeScript terminates.
func RegisterTerminateCallback(callbacks ...ExtNativescriptTerminateCallback) {
	terminateNativescriptCallbacks = append(terminateNativescriptCallbacks, callbacks...)
}
