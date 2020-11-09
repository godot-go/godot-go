package gdnative

import (
	"fmt"

	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

func NodePathField(key string, np NodePath) log.Field {
	return zap.String(key, fmt.Sprintf("NodePath(%s)", np.AsString()))
}

func Vector2Field(key string, vec2 Vector2) log.Field {
	return zap.String(key, fmt.Sprintf("vec2(%.2f,%.2f)", vec2.GetX(), vec2.GetY()))
}

func GodotObjectField(key string, obj *GodotObject) log.Field {
	return zap.String(key, obj.AddrAsString())
}

func StringField(key string, value string) log.Field {
	return zap.String(key, value)
}

func VariantField(key string, value Variant) log.Field {
	return zap.String(key, fmt.Sprintf("%d:%+v", value.GetType(), VariantToGoType(value)))
}

func TypeTagField(key string, typeTag TypeTag) log.Field {
	name := RegisterState.TagDB.GetRegisteredClassName(typeTag)

	return zap.String(key, name)
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
