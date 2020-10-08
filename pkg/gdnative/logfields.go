package gdnative

import (
	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

func GodotObjectField(key string, obj *GodotObject) log.Field {
	return zap.String(key, obj.AddrAsString())
}

func StringField(key string, value string) log.Field {
	return zap.String(key, value)
}

func TypeTagField(key string, typeTag TypeTag) log.Field {
	return zap.Uint(key, uint(typeTag))
}

func MethodTagField(key string, methodTag MethodTag) log.Field {
	return zap.Uint(key, uint(methodTag))
}

func NativeScriptClassField(key string, value NativeScriptClass) log.Field {
	return zap.Any(key, value)
}

func AnyField(key string, value interface{}) log.Field {
	return zap.Any(key, value)
}
