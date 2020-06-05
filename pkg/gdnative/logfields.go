package gdnative

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/sirupsen/logrus"
)

func WithRegisteredClassCB(tt TypeTag, base string, userData UserData) logrus.Fields {
	return logrus.Fields{
		"tt":       fmt.Sprintf("%d", uint32(tt)),
		"base":     base,
		"userData": fmt.Sprintf("%p", userData),
	}
}

func WithRegisteredClassFreeCB(tt TypeTag, base string) logrus.Fields {
	return logrus.Fields{
		"tt":   fmt.Sprintf("%d", uint32(tt)),
		"base": base,
	}
}

func WithRegisteredMethodCB(md MethodData, method string, args []reflect.Value) logrus.Fields {
	strArgs := make([]string, len(args))

	for i, a := range args {
		strArgs[i] = fmt.Sprintf("%v", a)
	}

	return logrus.Fields{
		"md":     fmt.Sprintf("%d", uint32(md)),
		"method": method,
		"args":   "[" + strings.Join(strArgs, ",") + "]",
	}
}

func WithRegisteredMethodFreeCB(md MethodData, base string) logrus.Fields {
	return logrus.Fields{
		"md":   fmt.Sprintf("%d", uint32(md)),
		"base": base,
	}
}

func WithVector2(v Vector2) logrus.Fields {
	return logrus.Fields{
		"vector2": fmt.Sprintf("vector2(%.2f, %.2f)", v.GetX(), v.GetY()),
	}
}

func WithRegisteredClass(name, base string) logrus.Fields {
	return logrus.Fields{
		"class": name,
		"base":  base,
	}
}

func WithObject(o *GodotObject) logrus.Fields {
	if o == nil {
		return logrus.Fields{
			"userData": "null",
		}
	} else {
		return logrus.Fields{
			"userData": o.AddrAsString(),
		}
	}
}

func WithUserData(userData UserData) logrus.Fields {
	return logrus.Fields{
		"userData": fmt.Sprintf("%p", userData),
	}
}

func WithTypeTag(tt TypeTag) logrus.Fields {
	name := RegisterState.TagDB.GetRegisteredClassName(tt)
	return logrus.Fields{
		"tt":       fmt.Sprintf("%d", uint32(tt)),
		"type_tag": name,
	}
}

func WithMethodTag(mt MethodTag) logrus.Fields {
	name := RegisterState.TagDB.GetRegisteredMethodName(mt)
	return logrus.Fields{
		"mt":         fmt.Sprintf("%d", uint32(mt)),
		"method_tag": name,
	}
}

func WithGodotString(key string, value String) logrus.Fields {
	return logrus.Fields{
		key: value.AsGoString(),
	}
}

func WithGodotStringName(key string, value StringName) logrus.Fields {
	n := value.GetName()
	return logrus.Fields{
		key: n.AsGoString(),
	}
}
