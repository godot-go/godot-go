package gdextension

//revive:disable

// #include <godot/gdextension_interface.h>
// #include <stdio.h>
// #include <stdlib.h>
// #include "gdextension_binding_init.h"
// #include "stacktrace.h"
import "C"
import (
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/gdextensionffi"
	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

type internalImpl struct {
	gdNativeInstances *SyncMap[ObjectID, GDExtensionClass]
	gdClassInstances  *SyncMap[GDObjectInstanceID, GDClass]
}

type GDExtensionBindingCallback func()

type GDExtensionClassGoConstructorFromOwner func(*GodotObject) GDExtensionClass

type GDClassGoConstructor func(data unsafe.Pointer) GDExtensionObjectPtr

var (
	internal internalImpl

	nullptr = unsafe.Pointer(nil)

	gdNativeConstructors                                  = NewSyncMap[string, GDExtensionClassGoConstructorFromOwner]()
	gdExtensionBindingGDExtensionInstanceBindingCallbacks = NewSyncMap[string, GDExtensionInstanceBindingCallbacks]()
	gdRegisteredGDClasses                                 = NewSyncMap[string, *ClassInfo]()
	gdExtensionBindingInitCallbacks                       [GDEXTENSION_MAX_INITIALIZATION_LEVEL]GDExtensionBindingCallback
	gdExtensionBindingTerminateCallbacks                  [GDEXTENSION_MAX_INITIALIZATION_LEVEL]GDExtensionBindingCallback
)

func _GDExtensionBindingInit(
	pGetProcAddress GDExtensionInterfaceGetProcAddress,
	pLibrary GDExtensionClassLibraryPtr,
	rInitialization *GDExtensionInitialization,
) bool {
	C.enablePrintStacktrace = log.GetLevel() == log.DebugLevel

	internal.gdNativeInstances = NewSyncMap[ObjectID, GDExtensionClass]()
	internal.gdClassInstances = NewSyncMap[GDObjectInstanceID, GDClass]()

	FFI.GetProcAddress = pGetProcAddress
	FFI.Library = pLibrary
	FFI.Token = unsafe.Pointer(&pLibrary)

	FFI.GetGodotVersion = (GDExtensionInterfaceGetGodotVersion)(LoadProcAddress("get_godot_version"))
	FFI.MemAlloc = (GDExtensionInterfaceMemAlloc)(LoadProcAddress("mem_alloc"))
	FFI.MemRealloc = (GDExtensionInterfaceMemRealloc)(LoadProcAddress("mem_realloc"))
	FFI.MemFree = (GDExtensionInterfaceMemFree)(LoadProcAddress("mem_free"))
	FFI.PrintError = (GDExtensionInterfacePrintError)(LoadProcAddress("print_error"))
	FFI.PrintErrorWithMessage = (GDExtensionInterfacePrintErrorWithMessage)(LoadProcAddress("print_error_with_message"))
	FFI.PrintWarning = (GDExtensionInterfacePrintWarning)(LoadProcAddress("print_warning"))
	FFI.PrintWarningWithMessage = (GDExtensionInterfacePrintWarningWithMessage)(LoadProcAddress("print_warning_with_message"))
	FFI.PrintScriptError = (GDExtensionInterfacePrintScriptError)(LoadProcAddress("print_script_error"))
	FFI.PrintScriptErrorWithMessage = (GDExtensionInterfacePrintScriptErrorWithMessage)(LoadProcAddress("print_script_error_with_message"))
	FFI.GetNativeStructSize = (GDExtensionInterfaceGetNativeStructSize)(LoadProcAddress("get_native_struct_size"))
	FFI.VariantNewCopy = (GDExtensionInterfaceVariantNewCopy)(LoadProcAddress("variant_new_copy"))
	FFI.VariantNewNil = (GDExtensionInterfaceVariantNewNil)(LoadProcAddress("variant_new_nil"))
	FFI.VariantDestroy = (GDExtensionInterfaceVariantDestroy)(LoadProcAddress("variant_destroy"))
	FFI.VariantCall = (GDExtensionInterfaceVariantCall)(LoadProcAddress("variant_call"))
	FFI.VariantCallStatic = (GDExtensionInterfaceVariantCallStatic)(LoadProcAddress("variant_call_static"))
	FFI.VariantEvaluate = (GDExtensionInterfaceVariantEvaluate)(LoadProcAddress("variant_evaluate"))
	FFI.VariantSet = (GDExtensionInterfaceVariantSet)(LoadProcAddress("variant_set"))
	FFI.VariantSetNamed = (GDExtensionInterfaceVariantSetNamed)(LoadProcAddress("variant_set_named"))
	FFI.VariantSetKeyed = (GDExtensionInterfaceVariantSetKeyed)(LoadProcAddress("variant_set_keyed"))
	FFI.VariantSetIndexed = (GDExtensionInterfaceVariantSetIndexed)(LoadProcAddress("variant_set_indexed"))
	FFI.VariantGet = (GDExtensionInterfaceVariantGet)(LoadProcAddress("variant_get"))
	FFI.VariantGetNamed = (GDExtensionInterfaceVariantGetNamed)(LoadProcAddress("variant_get_named"))
	FFI.VariantGetKeyed = (GDExtensionInterfaceVariantGetKeyed)(LoadProcAddress("variant_get_keyed"))
	FFI.VariantGetIndexed = (GDExtensionInterfaceVariantGetIndexed)(LoadProcAddress("variant_get_indexed"))
	FFI.VariantIterInit = (GDExtensionInterfaceVariantIterInit)(LoadProcAddress("variant_iter_init"))
	FFI.VariantIterNext = (GDExtensionInterfaceVariantIterNext)(LoadProcAddress("variant_iter_next"))
	FFI.VariantIterGet = (GDExtensionInterfaceVariantIterGet)(LoadProcAddress("variant_iter_get"))
	FFI.VariantHash = (GDExtensionInterfaceVariantHash)(LoadProcAddress("variant_hash"))
	FFI.VariantRecursiveHash = (GDExtensionInterfaceVariantRecursiveHash)(LoadProcAddress("variant_recursive_hash"))
	FFI.VariantHashCompare = (GDExtensionInterfaceVariantHashCompare)(LoadProcAddress("variant_hash_compare"))
	FFI.VariantBooleanize = (GDExtensionInterfaceVariantBooleanize)(LoadProcAddress("variant_booleanize"))
	FFI.VariantDuplicate = (GDExtensionInterfaceVariantDuplicate)(LoadProcAddress("variant_duplicate"))
	FFI.VariantStringify = (GDExtensionInterfaceVariantStringify)(LoadProcAddress("variant_stringify"))
	FFI.VariantGetType = (GDExtensionInterfaceVariantGetType)(LoadProcAddress("variant_get_type"))
	FFI.VariantHasMethod = (GDExtensionInterfaceVariantHasMethod)(LoadProcAddress("variant_has_method"))
	FFI.VariantHasMember = (GDExtensionInterfaceVariantHasMember)(LoadProcAddress("variant_has_member"))
	FFI.VariantHasKey = (GDExtensionInterfaceVariantHasKey)(LoadProcAddress("variant_has_key"))
	FFI.VariantGetTypeName = (GDExtensionInterfaceVariantGetTypeName)(LoadProcAddress("variant_get_type_name"))
	FFI.VariantCanConvert = (GDExtensionInterfaceVariantCanConvert)(LoadProcAddress("variant_can_convert"))
	FFI.VariantCanConvertStrict = (GDExtensionInterfaceVariantCanConvertStrict)(LoadProcAddress("variant_can_convert_strict"))
	FFI.GetVariantFromTypeConstructor = (GDExtensionInterfaceGetVariantFromTypeConstructor)(LoadProcAddress("get_variant_from_type_constructor"))
	FFI.GetVariantToTypeConstructor = (GDExtensionInterfaceGetVariantToTypeConstructor)(LoadProcAddress("get_variant_to_type_constructor"))
	FFI.VariantGetPtrOperatorEvaluator = (GDExtensionInterfaceVariantGetPtrOperatorEvaluator)(LoadProcAddress("variant_get_ptr_operator_evaluator"))
	FFI.VariantGetPtrBuiltinMethod = (GDExtensionInterfaceVariantGetPtrBuiltinMethod)(LoadProcAddress("variant_get_ptr_builtin_method"))
	FFI.VariantGetPtrConstructor = (GDExtensionInterfaceVariantGetPtrConstructor)(LoadProcAddress("variant_get_ptr_constructor"))
	FFI.VariantGetPtrDestructor = (GDExtensionInterfaceVariantGetPtrDestructor)(LoadProcAddress("variant_get_ptr_destructor"))
	FFI.VariantConstruct = (GDExtensionInterfaceVariantConstruct)(LoadProcAddress("variant_construct"))
	FFI.VariantGetPtrSetter = (GDExtensionInterfaceVariantGetPtrSetter)(LoadProcAddress("variant_get_ptr_setter"))
	FFI.VariantGetPtrGetter = (GDExtensionInterfaceVariantGetPtrGetter)(LoadProcAddress("variant_get_ptr_getter"))
	FFI.VariantGetPtrIndexedSetter = (GDExtensionInterfaceVariantGetPtrIndexedSetter)(LoadProcAddress("variant_get_ptr_indexed_setter"))
	FFI.VariantGetPtrIndexedGetter = (GDExtensionInterfaceVariantGetPtrIndexedGetter)(LoadProcAddress("variant_get_ptr_indexed_getter"))
	FFI.VariantGetPtrKeyedSetter = (GDExtensionInterfaceVariantGetPtrKeyedSetter)(LoadProcAddress("variant_get_ptr_keyed_setter"))
	FFI.VariantGetPtrKeyedGetter = (GDExtensionInterfaceVariantGetPtrKeyedGetter)(LoadProcAddress("variant_get_ptr_keyed_getter"))
	FFI.VariantGetPtrKeyedChecker = (GDExtensionInterfaceVariantGetPtrKeyedChecker)(LoadProcAddress("variant_get_ptr_keyed_checker"))
	FFI.VariantGetConstantValue = (GDExtensionInterfaceVariantGetConstantValue)(LoadProcAddress("variant_get_constant_value"))
	FFI.VariantGetPtrUtilityFunction = (GDExtensionInterfaceVariantGetPtrUtilityFunction)(LoadProcAddress("variant_get_ptr_utility_function"))
	FFI.StringNewWithLatin1Chars = (GDExtensionInterfaceStringNewWithLatin1Chars)(LoadProcAddress("string_new_with_latin1_chars"))
	FFI.StringNewWithUtf8Chars = (GDExtensionInterfaceStringNewWithUtf8Chars)(LoadProcAddress("string_new_with_utf8_chars"))
	FFI.StringNewWithUtf16Chars = (GDExtensionInterfaceStringNewWithUtf16Chars)(LoadProcAddress("string_new_with_utf16_chars"))
	FFI.StringNewWithUtf32Chars = (GDExtensionInterfaceStringNewWithUtf32Chars)(LoadProcAddress("string_new_with_utf32_chars"))
	FFI.StringNewWithWideChars = (GDExtensionInterfaceStringNewWithWideChars)(LoadProcAddress("string_new_with_wide_chars"))
	FFI.StringNewWithLatin1CharsAndLen = (GDExtensionInterfaceStringNewWithLatin1CharsAndLen)(LoadProcAddress("string_new_with_latin1_chars_and_len"))
	FFI.StringNewWithUtf8CharsAndLen = (GDExtensionInterfaceStringNewWithUtf8CharsAndLen)(LoadProcAddress("string_new_with_utf8_chars_and_len"))
	FFI.StringNewWithUtf16CharsAndLen = (GDExtensionInterfaceStringNewWithUtf16CharsAndLen)(LoadProcAddress("string_new_with_utf16_chars_and_len"))
	FFI.StringNewWithUtf32CharsAndLen = (GDExtensionInterfaceStringNewWithUtf32CharsAndLen)(LoadProcAddress("string_new_with_utf32_chars_and_len"))
	FFI.StringNewWithWideCharsAndLen = (GDExtensionInterfaceStringNewWithWideCharsAndLen)(LoadProcAddress("string_new_with_wide_chars_and_len"))
	FFI.StringToLatin1Chars = (GDExtensionInterfaceStringToLatin1Chars)(LoadProcAddress("string_to_latin1_chars"))
	FFI.StringToUtf8Chars = (GDExtensionInterfaceStringToUtf8Chars)(LoadProcAddress("string_to_utf8_chars"))
	FFI.StringToUtf16Chars = (GDExtensionInterfaceStringToUtf16Chars)(LoadProcAddress("string_to_utf16_chars"))
	FFI.StringToUtf32Chars = (GDExtensionInterfaceStringToUtf32Chars)(LoadProcAddress("string_to_utf32_chars"))
	FFI.StringToWideChars = (GDExtensionInterfaceStringToWideChars)(LoadProcAddress("string_to_wide_chars"))
	FFI.StringOperatorIndex = (GDExtensionInterfaceStringOperatorIndex)(LoadProcAddress("string_operator_index"))
	FFI.StringOperatorIndexConst = (GDExtensionInterfaceStringOperatorIndexConst)(LoadProcAddress("string_operator_index_const"))
	FFI.StringOperatorPlusEqString = (GDExtensionInterfaceStringOperatorPlusEqString)(LoadProcAddress("string_operator_plus_eq_string"))
	FFI.StringOperatorPlusEqChar = (GDExtensionInterfaceStringOperatorPlusEqChar)(LoadProcAddress("string_operator_plus_eq_char"))
	FFI.StringOperatorPlusEqCstr = (GDExtensionInterfaceStringOperatorPlusEqCstr)(LoadProcAddress("string_operator_plus_eq_cstr"))
	FFI.StringOperatorPlusEqWcstr = (GDExtensionInterfaceStringOperatorPlusEqWcstr)(LoadProcAddress("string_operator_plus_eq_wcstr"))
	FFI.StringOperatorPlusEqC32str = (GDExtensionInterfaceStringOperatorPlusEqC32str)(LoadProcAddress("string_operator_plus_eq_c32str"))
	FFI.XmlParserOpenBuffer = (GDExtensionInterfaceXmlParserOpenBuffer)(LoadProcAddress("xml_parser_open_buffer"))
	FFI.FileAccessStoreBuffer = (GDExtensionInterfaceFileAccessStoreBuffer)(LoadProcAddress("file_access_store_buffer"))
	FFI.FileAccessGetBuffer = (GDExtensionInterfaceFileAccessGetBuffer)(LoadProcAddress("file_access_get_buffer"))
	FFI.WorkerThreadPoolAddNativeGroupTask = (GDExtensionInterfaceWorkerThreadPoolAddNativeGroupTask)(LoadProcAddress("worker_thread_pool_add_native_group_task"))
	FFI.WorkerThreadPoolAddNativeTask = (GDExtensionInterfaceWorkerThreadPoolAddNativeTask)(LoadProcAddress("worker_thread_pool_add_native_task"))
	FFI.PackedByteArrayOperatorIndex = (GDExtensionInterfacePackedByteArrayOperatorIndex)(LoadProcAddress("packed_byte_array_operator_index"))
	FFI.PackedByteArrayOperatorIndexConst = (GDExtensionInterfacePackedByteArrayOperatorIndexConst)(LoadProcAddress("packed_byte_array_operator_index_const"))
	FFI.PackedColorArrayOperatorIndex = (GDExtensionInterfacePackedColorArrayOperatorIndex)(LoadProcAddress("packed_color_array_operator_index"))
	FFI.PackedColorArrayOperatorIndexConst = (GDExtensionInterfacePackedColorArrayOperatorIndexConst)(LoadProcAddress("packed_color_array_operator_index_const"))
	FFI.PackedFloat32ArrayOperatorIndex = (GDExtensionInterfacePackedFloat32ArrayOperatorIndex)(LoadProcAddress("packed_float32_array_operator_index"))
	FFI.PackedFloat32ArrayOperatorIndexConst = (GDExtensionInterfacePackedFloat32ArrayOperatorIndexConst)(LoadProcAddress("packed_float32_array_operator_index_const"))
	FFI.PackedFloat64ArrayOperatorIndex = (GDExtensionInterfacePackedFloat64ArrayOperatorIndex)(LoadProcAddress("packed_float64_array_operator_index"))
	FFI.PackedFloat64ArrayOperatorIndexConst = (GDExtensionInterfacePackedFloat64ArrayOperatorIndexConst)(LoadProcAddress("packed_float64_array_operator_index_const"))
	FFI.PackedInt32ArrayOperatorIndex = (GDExtensionInterfacePackedInt32ArrayOperatorIndex)(LoadProcAddress("packed_int32_array_operator_index"))
	FFI.PackedInt32ArrayOperatorIndexConst = (GDExtensionInterfacePackedInt32ArrayOperatorIndexConst)(LoadProcAddress("packed_int32_array_operator_index_const"))
	FFI.PackedInt64ArrayOperatorIndex = (GDExtensionInterfacePackedInt64ArrayOperatorIndex)(LoadProcAddress("packed_int64_array_operator_index"))
	FFI.PackedInt64ArrayOperatorIndexConst = (GDExtensionInterfacePackedInt64ArrayOperatorIndexConst)(LoadProcAddress("packed_int64_array_operator_index_const"))
	FFI.PackedStringArrayOperatorIndex = (GDExtensionInterfacePackedStringArrayOperatorIndex)(LoadProcAddress("packed_string_array_operator_index"))
	FFI.PackedStringArrayOperatorIndexConst = (GDExtensionInterfacePackedStringArrayOperatorIndexConst)(LoadProcAddress("packed_string_array_operator_index_const"))
	FFI.PackedVector2ArrayOperatorIndex = (GDExtensionInterfacePackedVector2ArrayOperatorIndex)(LoadProcAddress("packed_vector2_array_operator_index"))
	FFI.PackedVector2ArrayOperatorIndexConst = (GDExtensionInterfacePackedVector2ArrayOperatorIndexConst)(LoadProcAddress("packed_vector2_array_operator_index_const"))
	FFI.PackedVector3ArrayOperatorIndex = (GDExtensionInterfacePackedVector3ArrayOperatorIndex)(LoadProcAddress("packed_vector3_array_operator_index"))
	FFI.PackedVector3ArrayOperatorIndexConst = (GDExtensionInterfacePackedVector3ArrayOperatorIndexConst)(LoadProcAddress("packed_vector3_array_operator_index_const"))
	FFI.ArrayOperatorIndex = (GDExtensionInterfaceArrayOperatorIndex)(LoadProcAddress("array_operator_index"))
	FFI.ArrayOperatorIndexConst = (GDExtensionInterfaceArrayOperatorIndexConst)(LoadProcAddress("array_operator_index_const"))
	FFI.ArrayRef = (GDExtensionInterfaceArrayRef)(LoadProcAddress("array_ref"))
	FFI.ArraySetTyped = (GDExtensionInterfaceArraySetTyped)(LoadProcAddress("array_set_typed"))
	FFI.DictionaryOperatorIndex = (GDExtensionInterfaceDictionaryOperatorIndex)(LoadProcAddress("dictionary_operator_index"))
	FFI.DictionaryOperatorIndexConst = (GDExtensionInterfaceDictionaryOperatorIndexConst)(LoadProcAddress("dictionary_operator_index_const"))
	FFI.ObjectMethodBindCall = (GDExtensionInterfaceObjectMethodBindCall)(LoadProcAddress("object_method_bind_call"))
	FFI.ObjectMethodBindPtrcall = (GDExtensionInterfaceObjectMethodBindPtrcall)(LoadProcAddress("object_method_bind_ptrcall"))
	FFI.ObjectDestroy = (GDExtensionInterfaceObjectDestroy)(LoadProcAddress("object_destroy"))
	FFI.GlobalGetSingleton = (GDExtensionInterfaceGlobalGetSingleton)(LoadProcAddress("global_get_singleton"))
	FFI.ObjectGetInstanceBinding = (GDExtensionInterfaceObjectGetInstanceBinding)(LoadProcAddress("object_get_instance_binding"))
	FFI.ObjectSetInstanceBinding = (GDExtensionInterfaceObjectSetInstanceBinding)(LoadProcAddress("object_set_instance_binding"))
	FFI.ObjectSetInstance = (GDExtensionInterfaceObjectSetInstance)(LoadProcAddress("object_set_instance"))
	FFI.ObjectGetClassName = (GDExtensionInterfaceObjectGetClassName)(LoadProcAddress("object_get_class_name"))
	FFI.ObjectCastTo = (GDExtensionInterfaceObjectCastTo)(LoadProcAddress("object_cast_to"))
	FFI.ObjectGetInstanceFromId = (GDExtensionInterfaceObjectGetInstanceFromId)(LoadProcAddress("object_get_instance_from_id"))
	FFI.ObjectGetInstanceId = (GDExtensionInterfaceObjectGetInstanceId)(LoadProcAddress("object_get_instance_id"))
	FFI.RefGetObject = (GDExtensionInterfaceRefGetObject)(LoadProcAddress("ref_get_object"))
	FFI.RefSetObject = (GDExtensionInterfaceRefSetObject)(LoadProcAddress("ref_set_object"))
	FFI.ScriptInstanceCreate = (GDExtensionInterfaceScriptInstanceCreate)(LoadProcAddress("script_instance_create"))
	FFI.ClassdbConstructObject = (GDExtensionInterfaceClassdbConstructObject)(LoadProcAddress("classdb_construct_object"))
	FFI.ClassdbGetMethodBind = (GDExtensionInterfaceClassdbGetMethodBind)(LoadProcAddress("classdb_get_method_bind"))
	FFI.ClassdbGetClassTag = (GDExtensionInterfaceClassdbGetClassTag)(LoadProcAddress("classdb_get_class_tag"))
	FFI.ClassdbRegisterExtensionClass = (GDExtensionInterfaceClassdbRegisterExtensionClass)(LoadProcAddress("classdb_register_extension_class"))
	FFI.ClassdbRegisterExtensionClassMethod = (GDExtensionInterfaceClassdbRegisterExtensionClassMethod)(LoadProcAddress("classdb_register_extension_class_method"))
	FFI.ClassdbRegisterExtensionClassIntegerConstant = (GDExtensionInterfaceClassdbRegisterExtensionClassIntegerConstant)(LoadProcAddress("classdb_register_extension_class_integer_constant"))
	FFI.ClassdbRegisterExtensionClassProperty = (GDExtensionInterfaceClassdbRegisterExtensionClassProperty)(LoadProcAddress("classdb_register_extension_class_property"))
	FFI.ClassdbRegisterExtensionClassPropertyGroup = (GDExtensionInterfaceClassdbRegisterExtensionClassPropertyGroup)(LoadProcAddress("classdb_register_extension_class_property_group"))
	FFI.ClassdbRegisterExtensionClassPropertySubgroup = (GDExtensionInterfaceClassdbRegisterExtensionClassPropertySubgroup)(LoadProcAddress("classdb_register_extension_class_property_subgroup"))
	FFI.ClassdbRegisterExtensionClassSignal = (GDExtensionInterfaceClassdbRegisterExtensionClassSignal)(LoadProcAddress("classdb_register_extension_class_signal"))
	FFI.ClassdbUnregisterExtensionClass = (GDExtensionInterfaceClassdbUnregisterExtensionClass)(LoadProcAddress("classdb_unregister_extension_class"))
	FFI.GetLibraryPath = (GDExtensionInterfaceGetLibraryPath)(LoadProcAddress("get_library_path"))
	FFI.EditorAddPlugin = (GDExtensionInterfaceEditorAddPlugin)(LoadProcAddress("editor_add_plugin"))
	FFI.EditorRemovePlugin = (GDExtensionInterfaceEditorRemovePlugin)(LoadProcAddress("editor_remove_plugin"))

	// Load the Godot version.
	CallFunc_GDExtensionInterfaceGetGodotVersion(&FFI.GodotVersion)

	log.Info("godot version",
		zap.Int32("major", FFI.GodotVersion.GetMajor()),
		zap.Int32("minor", FFI.GodotVersion.GetMinor()),
	)

	rInitialization.SetCallbacks(
		(*[0]byte)(C.cgo_callfn_GDExtensionBindingInitializeLevel),
		(*[0]byte)(C.cgo_callfn_GDExtensionBindingDeinitializeLevel),
	)

	var hasInit bool

	for i := GDExtensionInitializationLevel(0); i < GDEXTENSION_MAX_INITIALIZATION_LEVEL; i++ {
		if gdExtensionBindingInitCallbacks[i] != nil {
			rInitialization.SetInitializationLevel(i)
			hasInit = true
			break
		}
	}

	if !hasInit {
		panic("At least one initialization callback must be defined.")
	}

	variantInitBindings()

	return true
}

//export GDExtensionBindingInitializeLevel
func GDExtensionBindingInitializeLevel(userdata unsafe.Pointer, pLevel C.GDExtensionInitializationLevel) {
	classdbCurrentLevel = (GDExtensionInitializationLevel)(pLevel)

	if fn := gdExtensionBindingInitCallbacks[pLevel]; fn != nil {
		log.Debug("GDExtensionBindingInitializeLevel init", zap.Int32("level", (int32)(pLevel)))
		fn()
	}

	classDBInitialize(classdbCurrentLevel)
}

//export GDExtensionBindingDeinitializeLevel
func GDExtensionBindingDeinitializeLevel(userdata unsafe.Pointer, pLevel C.GDExtensionInitializationLevel) {
	classdbCurrentLevel = (GDExtensionInitializationLevel)(pLevel)
	classDBDeinitialize(classdbCurrentLevel)

	if gdExtensionBindingTerminateCallbacks[pLevel] != nil {
		gdExtensionBindingTerminateCallbacks[pLevel]()
	}
}

func GDExtensionBindingCreateInstanceCallback(pToken unsafe.Pointer, pInstance unsafe.Pointer) Wrapped {
	if pToken != unsafe.Pointer(FFI.Library) {
		panic("Asking for creating instance with invalid token.")
	}

	owner := (*GodotObject)(pInstance)

	id := CallFunc_GDExtensionInterfaceObjectGetInstanceId((GDExtensionConstObjectPtr)(owner))

	log.Debug("GDExtensionBindingCreateInstanceCallback called", zap.Any("id", id))

	obj := NewGDExtensionClassFromObjectOwner(owner).(Object)

	strClass := obj.GetClass()

	cn := strClass.ToAscii()

	w := obj.CastTo(cn)
	return w
}

func GDExtensionBindingFreeInstanceCallback(pToken unsafe.Pointer, pInstance unsafe.Pointer, pBinding unsafe.Pointer) {
	if pToken != unsafe.Pointer(FFI.Library) {
		panic("Asking for freeing instance with invalid token.")
	}

	w := (*WrappedImpl)(pBinding)

	CallFunc_GDExtensionInterfaceObjectDestroy((GDExtensionObjectPtr)(w.Owner))
}

type InitObject struct {
	getProcAddress GDExtensionInterfaceGetProcAddress
	library        GDExtensionClassLibraryPtr
	initialization *GDExtensionInitialization
}

func (o InitObject) RegisterCoreInitializer(pCoreInit GDExtensionBindingCallback) {
	gdExtensionBindingInitCallbacks[GDEXTENSION_INITIALIZATION_CORE] = pCoreInit
}

func (o InitObject) RegisterServerInitializer(pServerInit GDExtensionBindingCallback) {
	gdExtensionBindingInitCallbacks[GDEXTENSION_INITIALIZATION_SERVERS] = pServerInit
}

func (o InitObject) RegisterSceneInitializer(pSceneInit GDExtensionBindingCallback) {
	gdExtensionBindingInitCallbacks[GDEXTENSION_INITIALIZATION_SCENE] = pSceneInit
}

func (o InitObject) RegisterEditorInitializer(pEditorInit GDExtensionBindingCallback) {
	gdExtensionBindingInitCallbacks[GDEXTENSION_INITIALIZATION_EDITOR] = pEditorInit
}

func (o InitObject) RegisterCoreTerminator(pCoreTerminate GDExtensionBindingCallback) {
	gdExtensionBindingTerminateCallbacks[GDEXTENSION_INITIALIZATION_CORE] = pCoreTerminate
}

func (o InitObject) RegisterServerTerminator(pServerTerminate GDExtensionBindingCallback) {
	gdExtensionBindingTerminateCallbacks[GDEXTENSION_INITIALIZATION_SERVERS] = pServerTerminate
}

func (o InitObject) RegisterSceneTerminator(pSceneTerminate GDExtensionBindingCallback) {
	gdExtensionBindingTerminateCallbacks[GDEXTENSION_INITIALIZATION_SCENE] = pSceneTerminate
}

func (o InitObject) RegisterEditorTerminator(pEditorTerminate GDExtensionBindingCallback) {
	gdExtensionBindingTerminateCallbacks[GDEXTENSION_INITIALIZATION_EDITOR] = pEditorTerminate
}

func (o InitObject) Init() bool {
	return _GDExtensionBindingInit(o.getProcAddress, o.library, o.initialization)
}

func NewInitObject(
	getProcAddress GDExtensionInterfaceGetProcAddress,
	library GDExtensionClassLibraryPtr,
	initialization *GDExtensionInitialization,
) *InitObject {
	return &InitObject{
		getProcAddress: getProcAddress,
		library:        library,
		initialization: initialization,
	}
}
