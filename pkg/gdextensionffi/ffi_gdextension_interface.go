package gdextensionffi

//revive:disable

import (
	"fmt"
	"unsafe"
)

type GDExtensionInterface struct {
	GetProcAddress GDExtensionInterfaceGetProcAddress
	Library        GDExtensionClassLibraryPtr
	Token          unsafe.Pointer

	GodotVersion GDExtensionGodotVersion

	// All of the GDExtension interface functions.
	GetGodotVersion                               GDExtensionInterfaceGetGodotVersion
	MemAlloc                                      GDExtensionInterfaceMemAlloc
	MemRealloc                                    GDExtensionInterfaceMemRealloc
	MemFree                                       GDExtensionInterfaceMemFree
	PrintError                                    GDExtensionInterfacePrintError
	PrintErrorWithMessage                         GDExtensionInterfacePrintErrorWithMessage
	PrintWarning                                  GDExtensionInterfacePrintWarning
	PrintWarningWithMessage                       GDExtensionInterfacePrintWarningWithMessage
	PrintScriptError                              GDExtensionInterfacePrintScriptError
	PrintScriptErrorWithMessage                   GDExtensionInterfacePrintScriptErrorWithMessage
	GetNativeStructSize                           GDExtensionInterfaceGetNativeStructSize
	VariantNewCopy                                GDExtensionInterfaceVariantNewCopy
	VariantNewNil                                 GDExtensionInterfaceVariantNewNil
	VariantDestroy                                GDExtensionInterfaceVariantDestroy
	VariantCall                                   GDExtensionInterfaceVariantCall
	VariantCallStatic                             GDExtensionInterfaceVariantCallStatic
	VariantEvaluate                               GDExtensionInterfaceVariantEvaluate
	VariantSet                                    GDExtensionInterfaceVariantSet
	VariantSetNamed                               GDExtensionInterfaceVariantSetNamed
	VariantSetKeyed                               GDExtensionInterfaceVariantSetKeyed
	VariantSetIndexed                             GDExtensionInterfaceVariantSetIndexed
	VariantGet                                    GDExtensionInterfaceVariantGet
	VariantGetNamed                               GDExtensionInterfaceVariantGetNamed
	VariantGetKeyed                               GDExtensionInterfaceVariantGetKeyed
	VariantGetIndexed                             GDExtensionInterfaceVariantGetIndexed
	VariantIterInit                               GDExtensionInterfaceVariantIterInit
	VariantIterNext                               GDExtensionInterfaceVariantIterNext
	VariantIterGet                                GDExtensionInterfaceVariantIterGet
	VariantHash                                   GDExtensionInterfaceVariantHash
	VariantRecursiveHash                          GDExtensionInterfaceVariantRecursiveHash
	VariantHashCompare                            GDExtensionInterfaceVariantHashCompare
	VariantBooleanize                             GDExtensionInterfaceVariantBooleanize
	VariantDuplicate                              GDExtensionInterfaceVariantDuplicate
	VariantStringify                              GDExtensionInterfaceVariantStringify
	VariantGetType                                GDExtensionInterfaceVariantGetType
	VariantHasMethod                              GDExtensionInterfaceVariantHasMethod
	VariantHasMember                              GDExtensionInterfaceVariantHasMember
	VariantHasKey                                 GDExtensionInterfaceVariantHasKey
	VariantGetTypeName                            GDExtensionInterfaceVariantGetTypeName
	VariantCanConvert                             GDExtensionInterfaceVariantCanConvert
	VariantCanConvertStrict                       GDExtensionInterfaceVariantCanConvertStrict
	GetVariantFromTypeConstructor                 GDExtensionInterfaceGetVariantFromTypeConstructor
	GetVariantToTypeConstructor                   GDExtensionInterfaceGetVariantToTypeConstructor
	VariantGetPtrOperatorEvaluator                GDExtensionInterfaceVariantGetPtrOperatorEvaluator
	VariantGetPtrBuiltinMethod                    GDExtensionInterfaceVariantGetPtrBuiltinMethod
	VariantGetPtrConstructor                      GDExtensionInterfaceVariantGetPtrConstructor
	VariantGetPtrDestructor                       GDExtensionInterfaceVariantGetPtrDestructor
	VariantConstruct                              GDExtensionInterfaceVariantConstruct
	VariantGetPtrSetter                           GDExtensionInterfaceVariantGetPtrSetter
	VariantGetPtrGetter                           GDExtensionInterfaceVariantGetPtrGetter
	VariantGetPtrIndexedSetter                    GDExtensionInterfaceVariantGetPtrIndexedSetter
	VariantGetPtrIndexedGetter                    GDExtensionInterfaceVariantGetPtrIndexedGetter
	VariantGetPtrKeyedSetter                      GDExtensionInterfaceVariantGetPtrKeyedSetter
	VariantGetPtrKeyedGetter                      GDExtensionInterfaceVariantGetPtrKeyedGetter
	VariantGetPtrKeyedChecker                     GDExtensionInterfaceVariantGetPtrKeyedChecker
	VariantGetConstantValue                       GDExtensionInterfaceVariantGetConstantValue
	VariantGetPtrUtilityFunction                  GDExtensionInterfaceVariantGetPtrUtilityFunction
	StringNewWithLatin1Chars                      GDExtensionInterfaceStringNewWithLatin1Chars
	StringNewWithUtf8Chars                        GDExtensionInterfaceStringNewWithUtf8Chars
	StringNewWithUtf16Chars                       GDExtensionInterfaceStringNewWithUtf16Chars
	StringNewWithUtf32Chars                       GDExtensionInterfaceStringNewWithUtf32Chars
	StringNewWithWideChars                        GDExtensionInterfaceStringNewWithWideChars
	StringNewWithLatin1CharsAndLen                GDExtensionInterfaceStringNewWithLatin1CharsAndLen
	StringNewWithUtf8CharsAndLen                  GDExtensionInterfaceStringNewWithUtf8CharsAndLen
	StringNewWithUtf16CharsAndLen                 GDExtensionInterfaceStringNewWithUtf16CharsAndLen
	StringNewWithUtf32CharsAndLen                 GDExtensionInterfaceStringNewWithUtf32CharsAndLen
	StringNewWithWideCharsAndLen                  GDExtensionInterfaceStringNewWithWideCharsAndLen
	StringResize                                  GDExtensionInterfaceStringResize
	StringToLatin1Chars                           GDExtensionInterfaceStringToLatin1Chars
	StringToUtf8Chars                             GDExtensionInterfaceStringToUtf8Chars
	StringToUtf16Chars                            GDExtensionInterfaceStringToUtf16Chars
	StringToUtf32Chars                            GDExtensionInterfaceStringToUtf32Chars
	StringToWideChars                             GDExtensionInterfaceStringToWideChars
	StringOperatorIndex                           GDExtensionInterfaceStringOperatorIndex
	StringOperatorIndexConst                      GDExtensionInterfaceStringOperatorIndexConst
	StringOperatorPlusEqString                    GDExtensionInterfaceStringOperatorPlusEqString
	StringOperatorPlusEqChar                      GDExtensionInterfaceStringOperatorPlusEqChar
	StringOperatorPlusEqCstr                      GDExtensionInterfaceStringOperatorPlusEqCstr
	StringOperatorPlusEqWcstr                     GDExtensionInterfaceStringOperatorPlusEqWcstr
	StringOperatorPlusEqC32str                    GDExtensionInterfaceStringOperatorPlusEqC32str
	XmlParserOpenBuffer                           GDExtensionInterfaceXmlParserOpenBuffer
	FileAccessStoreBuffer                         GDExtensionInterfaceFileAccessStoreBuffer
	FileAccessGetBuffer                           GDExtensionInterfaceFileAccessGetBuffer
	WorkerThreadPoolAddNativeGroupTask            GDExtensionInterfaceWorkerThreadPoolAddNativeGroupTask
	WorkerThreadPoolAddNativeTask                 GDExtensionInterfaceWorkerThreadPoolAddNativeTask
	PackedByteArrayOperatorIndex                  GDExtensionInterfacePackedByteArrayOperatorIndex
	PackedByteArrayOperatorIndexConst             GDExtensionInterfacePackedByteArrayOperatorIndexConst
	PackedColorArrayOperatorIndex                 GDExtensionInterfacePackedColorArrayOperatorIndex
	PackedColorArrayOperatorIndexConst            GDExtensionInterfacePackedColorArrayOperatorIndexConst
	PackedFloat32ArrayOperatorIndex               GDExtensionInterfacePackedFloat32ArrayOperatorIndex
	PackedFloat32ArrayOperatorIndexConst          GDExtensionInterfacePackedFloat32ArrayOperatorIndexConst
	PackedFloat64ArrayOperatorIndex               GDExtensionInterfacePackedFloat64ArrayOperatorIndex
	PackedFloat64ArrayOperatorIndexConst          GDExtensionInterfacePackedFloat64ArrayOperatorIndexConst
	PackedInt32ArrayOperatorIndex                 GDExtensionInterfacePackedInt32ArrayOperatorIndex
	PackedInt32ArrayOperatorIndexConst            GDExtensionInterfacePackedInt32ArrayOperatorIndexConst
	PackedInt64ArrayOperatorIndex                 GDExtensionInterfacePackedInt64ArrayOperatorIndex
	PackedInt64ArrayOperatorIndexConst            GDExtensionInterfacePackedInt64ArrayOperatorIndexConst
	PackedStringArrayOperatorIndex                GDExtensionInterfacePackedStringArrayOperatorIndex
	PackedStringArrayOperatorIndexConst           GDExtensionInterfacePackedStringArrayOperatorIndexConst
	PackedVector2ArrayOperatorIndex               GDExtensionInterfacePackedVector2ArrayOperatorIndex
	PackedVector2ArrayOperatorIndexConst          GDExtensionInterfacePackedVector2ArrayOperatorIndexConst
	PackedVector3ArrayOperatorIndex               GDExtensionInterfacePackedVector3ArrayOperatorIndex
	PackedVector3ArrayOperatorIndexConst          GDExtensionInterfacePackedVector3ArrayOperatorIndexConst
	ArrayOperatorIndex                            GDExtensionInterfaceArrayOperatorIndex
	ArrayOperatorIndexConst                       GDExtensionInterfaceArrayOperatorIndexConst
	ArrayRef                                      GDExtensionInterfaceArrayRef
	ArraySetTyped                                 GDExtensionInterfaceArraySetTyped
	DictionaryOperatorIndex                       GDExtensionInterfaceDictionaryOperatorIndex
	DictionaryOperatorIndexConst                  GDExtensionInterfaceDictionaryOperatorIndexConst
	ObjectMethodBindCall                          GDExtensionInterfaceObjectMethodBindCall
	ObjectMethodBindPtrcall                       GDExtensionInterfaceObjectMethodBindPtrcall
	ObjectDestroy                                 GDExtensionInterfaceObjectDestroy
	GlobalGetSingleton                            GDExtensionInterfaceGlobalGetSingleton
	ObjectGetInstanceBinding                      GDExtensionInterfaceObjectGetInstanceBinding
	ObjectSetInstanceBinding                      GDExtensionInterfaceObjectSetInstanceBinding
	ObjectSetInstance                             GDExtensionInterfaceObjectSetInstance
	ObjectGetClassName                            GDExtensionInterfaceObjectGetClassName
	ObjectCastTo                                  GDExtensionInterfaceObjectCastTo
	ObjectGetInstanceFromId                       GDExtensionInterfaceObjectGetInstanceFromId
	ObjectGetInstanceId                           GDExtensionInterfaceObjectGetInstanceId
	ObjectGetScriptInstance                       GDExtensionInterfaceObjectGetScriptInstance
	RefGetObject                                  GDExtensionInterfaceRefGetObject
	RefSetObject                                  GDExtensionInterfaceRefSetObject
	ScriptInstanceCreate                          GDExtensionInterfaceScriptInstanceCreate
	ClassdbConstructObject                        GDExtensionInterfaceClassdbConstructObject
	ClassdbGetMethodBind                          GDExtensionInterfaceClassdbGetMethodBind
	ClassdbGetClassTag                            GDExtensionInterfaceClassdbGetClassTag
	ClassdbRegisterExtensionClass                 GDExtensionInterfaceClassdbRegisterExtensionClass
	ClassdbRegisterExtensionClassMethod           GDExtensionInterfaceClassdbRegisterExtensionClassMethod
	ClassdbRegisterExtensionClassIntegerConstant  GDExtensionInterfaceClassdbRegisterExtensionClassIntegerConstant
	ClassdbRegisterExtensionClassProperty         GDExtensionInterfaceClassdbRegisterExtensionClassProperty
	ClassdbRegisterExtensionClassPropertyGroup    GDExtensionInterfaceClassdbRegisterExtensionClassPropertyGroup
	ClassdbRegisterExtensionClassPropertyIndexed  GDExtensionInterfaceClassdbRegisterExtensionClassPropertyIndexed
	ClassdbRegisterExtensionClassPropertySubgroup GDExtensionInterfaceClassdbRegisterExtensionClassPropertySubgroup
	ClassdbRegisterExtensionClassSignal           GDExtensionInterfaceClassdbRegisterExtensionClassSignal
	ClassdbUnregisterExtensionClass               GDExtensionInterfaceClassdbUnregisterExtensionClass
	GetLibraryPath                                GDExtensionInterfaceGetLibraryPath
	EditorAddPlugin                               GDExtensionInterfaceEditorAddPlugin
	EditorRemovePlugin                            GDExtensionInterfaceEditorRemovePlugin
}

var (
	FFI GDExtensionInterface
)

func LoadProcAddress(funcName string) unsafe.Pointer {
	ret := CallFunc_GDExtensionInterfaceGetProcAddress(funcName)
	if ret == nil {
		panic(fmt.Sprintf("Unable to load GDExtension interface function %s()", funcName))
	}

	return unsafe.Pointer(ret)
}

func (gv GDExtensionGodotVersion) GetMajor() int32 {
	return int32(gv.major)
}

func (gv GDExtensionGodotVersion) GetMinor() int32 {
	return int32(gv.minor)
}