package gdclassimpl

import (
	"testing"

	"github.com/godot-go/godot-go/cmd/extensionapiparser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeriveSingletonAccessorsNameTypeDistinct(t *testing.T) {
	eapi := extensionapiparser.ExtensionApi{
		Singletons: []extensionapiparser.Singleton{
			{Name: "OS", Type: "OS"},
			{Name: "TheVoice", Type: "Engine"},
		},
	}

	infos := DeriveSingletonAccessors(eapi)

	require.Len(t, infos, 2)

	assert.Equal(t, SingletonAccessorInfo{
		Singleton:       "OS",
		ClassName:       "OS",
		FuncName:        "GetOSSingleton",
		BuilderName:     "buildOSSingleton",
		SlotVar:         "osSingletonSlot",
		LockVar:         "osSingletonLock",
		ConstructorName: "NewOSWithGodotOwnerObject",
	}, infos[0])

	// The accessor name follows the singleton name while the type and the
	// constructor follow the class type: name and type must not be conflated.
	assert.Equal(t, "GetTheVoiceSingleton", infos[1].FuncName)
	assert.Equal(t, "Engine", infos[1].ClassName)
	assert.Equal(t, "NewEngineWithGodotOwnerObject", infos[1].ConstructorName)
}

func TestDeriveSingletonAccessorsCoversRealApi(t *testing.T) {
	eapi, err := extensionapiparser.ParseExtensionApiJson(projectRoot)
	require.NoError(t, err)
	require.NotEmpty(t, eapi.Singletons)

	infos := DeriveSingletonAccessors(eapi)

	require.Len(t, infos, len(eapi.Singletons))

	seenFuncs := map[string]bool{}
	seenVars := map[string]bool{}
	for _, info := range infos {
		assert.False(t, seenFuncs[info.FuncName], "duplicate accessor %q", info.FuncName)
		seenFuncs[info.FuncName] = true
		for _, ident := range []string{info.SlotVar, info.LockVar} {
			assert.False(t, seenVars[ident], "duplicate identifier %q", ident)
			seenVars[ident] = true
		}
	}

	// Every class-typed entry must exist in the API classes (each class gets a
	// New<Class>WithGodotOwnerObject constructor), otherwise the generated
	// file cannot compile.
	classNames := map[string]bool{}
	for _, class := range eapi.Classes {
		classNames[class.Name] = true
	}
	for _, info := range infos {
		assert.True(t, classNames[info.ClassName],
			"singleton %q references unknown class %q", info.Singleton, info.ClassName)
	}
}
