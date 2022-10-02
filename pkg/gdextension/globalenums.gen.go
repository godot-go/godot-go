package gdextension

/*------------------------------------------------------------------------------
//   This code was generated by template globalenums.go.tmpl.
//
//   Changes to this file may cause incorrect behavior and will be lost if
//   the code is regenerated. Any updates should be done in
//   "globalenums.go.tmpl" so they can be included in the generated
//   code.
//----------------------------------------------------------------------------*/

//revive:disable

// #include <godot/gdnative_interface.h>
// #include <stdio.h>
// #include <stdlib.h>
import "C"

type Side int
const (
	SIDE_LEFT Side = 0
	SIDE_TOP  = 1
	SIDE_RIGHT  = 2
	SIDE_BOTTOM  = 3
	)
type Corner int
const (
	CORNER_TOP_LEFT Corner = 0
	CORNER_TOP_RIGHT  = 1
	CORNER_BOTTOM_RIGHT  = 2
	CORNER_BOTTOM_LEFT  = 3
	)
type Orientation int
const (
	VERTICAL Orientation = 1
	HORIZONTAL  = 0
	)
type ClockDirection int
const (
	CLOCKWISE ClockDirection = 0
	COUNTERCLOCKWISE  = 1
	)
type HorizontalAlignment int
const (
	HORIZONTAL_ALIGNMENT_LEFT HorizontalAlignment = 0
	HORIZONTAL_ALIGNMENT_CENTER  = 1
	HORIZONTAL_ALIGNMENT_RIGHT  = 2
	HORIZONTAL_ALIGNMENT_FILL  = 3
	)
type VerticalAlignment int
const (
	VERTICAL_ALIGNMENT_TOP VerticalAlignment = 0
	VERTICAL_ALIGNMENT_CENTER  = 1
	VERTICAL_ALIGNMENT_BOTTOM  = 2
	VERTICAL_ALIGNMENT_FILL  = 3
	)
type InlineAlignment int
const (
	INLINE_ALIGNMENT_TOP_TO InlineAlignment = 0
	INLINE_ALIGNMENT_CENTER_TO  = 1
	INLINE_ALIGNMENT_BOTTOM_TO  = 2
	INLINE_ALIGNMENT_TO_TOP  = 0
	INLINE_ALIGNMENT_TO_CENTER  = 4
	INLINE_ALIGNMENT_TO_BASELINE  = 8
	INLINE_ALIGNMENT_TO_BOTTOM  = 12
	INLINE_ALIGNMENT_TOP  = 0
	INLINE_ALIGNMENT_CENTER  = 5
	INLINE_ALIGNMENT_BOTTOM  = 14
	INLINE_ALIGNMENT_IMAGE_MASK  = 3
	INLINE_ALIGNMENT_TEXT_MASK  = 12
	)
type Key int
const (
	KEY_NONE Key = 0
	KEY_SPECIAL  = 4194304
	KEY_ESCAPE  = 4194305
	KEY_TAB  = 4194306
	KEY_BACKTAB  = 4194307
	KEY_BACKSPACE  = 4194308
	KEY_ENTER  = 4194309
	KEY_KP_ENTER  = 4194310
	KEY_INSERT  = 4194311
	KEY_DELETE  = 4194312
	KEY_PAUSE  = 4194313
	KEY_PRINT  = 4194314
	KEY_SYSREQ  = 4194315
	KEY_CLEAR  = 4194316
	KEY_HOME  = 4194317
	KEY_END  = 4194318
	KEY_LEFT  = 4194319
	KEY_UP  = 4194320
	KEY_RIGHT  = 4194321
	KEY_DOWN  = 4194322
	KEY_PAGEUP  = 4194323
	KEY_PAGEDOWN  = 4194324
	KEY_SHIFT  = 4194325
	KEY_CTRL  = 4194326
	KEY_META  = 4194327
	KEY_ALT  = 4194328
	KEY_CAPSLOCK  = 4194329
	KEY_NUMLOCK  = 4194330
	KEY_SCROLLLOCK  = 4194331
	KEY_F1  = 4194332
	KEY_F2  = 4194333
	KEY_F3  = 4194334
	KEY_F4  = 4194335
	KEY_F5  = 4194336
	KEY_F6  = 4194337
	KEY_F7  = 4194338
	KEY_F8  = 4194339
	KEY_F9  = 4194340
	KEY_F10  = 4194341
	KEY_F11  = 4194342
	KEY_F12  = 4194343
	KEY_F13  = 4194344
	KEY_F14  = 4194345
	KEY_F15  = 4194346
	KEY_F16  = 4194347
	KEY_F17  = 4194348
	KEY_F18  = 4194349
	KEY_F19  = 4194350
	KEY_F20  = 4194351
	KEY_F21  = 4194352
	KEY_F22  = 4194353
	KEY_F23  = 4194354
	KEY_F24  = 4194355
	KEY_F25  = 4194356
	KEY_F26  = 4194357
	KEY_F27  = 4194358
	KEY_F28  = 4194359
	KEY_F29  = 4194360
	KEY_F30  = 4194361
	KEY_F31  = 4194362
	KEY_F32  = 4194363
	KEY_F33  = 4194364
	KEY_F34  = 4194365
	KEY_F35  = 4194366
	KEY_KP_MULTIPLY  = 4194433
	KEY_KP_DIVIDE  = 4194434
	KEY_KP_SUBTRACT  = 4194435
	KEY_KP_PERIOD  = 4194436
	KEY_KP_ADD  = 4194437
	KEY_KP_0  = 4194438
	KEY_KP_1  = 4194439
	KEY_KP_2  = 4194440
	KEY_KP_3  = 4194441
	KEY_KP_4  = 4194442
	KEY_KP_5  = 4194443
	KEY_KP_6  = 4194444
	KEY_KP_7  = 4194445
	KEY_KP_8  = 4194446
	KEY_KP_9  = 4194447
	KEY_SUPER_L  = 4194368
	KEY_SUPER_R  = 4194369
	KEY_MENU  = 4194370
	KEY_HYPER_L  = 4194371
	KEY_HYPER_R  = 4194372
	KEY_HELP  = 4194373
	KEY_DIRECTION_L  = 4194374
	KEY_DIRECTION_R  = 4194375
	KEY_BACK  = 4194376
	KEY_FORWARD  = 4194377
	KEY_STOP  = 4194378
	KEY_REFRESH  = 4194379
	KEY_VOLUMEDOWN  = 4194380
	KEY_VOLUMEMUTE  = 4194381
	KEY_VOLUMEUP  = 4194382
	KEY_BASSBOOST  = 4194383
	KEY_BASSUP  = 4194384
	KEY_BASSDOWN  = 4194385
	KEY_TREBLEUP  = 4194386
	KEY_TREBLEDOWN  = 4194387
	KEY_MEDIAPLAY  = 4194388
	KEY_MEDIASTOP  = 4194389
	KEY_MEDIAPREVIOUS  = 4194390
	KEY_MEDIANEXT  = 4194391
	KEY_MEDIARECORD  = 4194392
	KEY_HOMEPAGE  = 4194393
	KEY_FAVORITES  = 4194394
	KEY_SEARCH  = 4194395
	KEY_STANDBY  = 4194396
	KEY_OPENURL  = 4194397
	KEY_LAUNCHMAIL  = 4194398
	KEY_LAUNCHMEDIA  = 4194399
	KEY_LAUNCH0  = 4194400
	KEY_LAUNCH1  = 4194401
	KEY_LAUNCH2  = 4194402
	KEY_LAUNCH3  = 4194403
	KEY_LAUNCH4  = 4194404
	KEY_LAUNCH5  = 4194405
	KEY_LAUNCH6  = 4194406
	KEY_LAUNCH7  = 4194407
	KEY_LAUNCH8  = 4194408
	KEY_LAUNCH9  = 4194409
	KEY_LAUNCHA  = 4194410
	KEY_LAUNCHB  = 4194411
	KEY_LAUNCHC  = 4194412
	KEY_LAUNCHD  = 4194413
	KEY_LAUNCHE  = 4194414
	KEY_LAUNCHF  = 4194415
	KEY_UNKNOWN  = 16777215
	KEY_SPACE  = 32
	KEY_EXCLAM  = 33
	KEY_QUOTEDBL  = 34
	KEY_NUMBERSIGN  = 35
	KEY_DOLLAR  = 36
	KEY_PERCENT  = 37
	KEY_AMPERSAND  = 38
	KEY_APOSTROPHE  = 39
	KEY_PARENLEFT  = 40
	KEY_PARENRIGHT  = 41
	KEY_ASTERISK  = 42
	KEY_PLUS  = 43
	KEY_COMMA  = 44
	KEY_MINUS  = 45
	KEY_PERIOD  = 46
	KEY_SLASH  = 47
	KEY_0  = 48
	KEY_1  = 49
	KEY_2  = 50
	KEY_3  = 51
	KEY_4  = 52
	KEY_5  = 53
	KEY_6  = 54
	KEY_7  = 55
	KEY_8  = 56
	KEY_9  = 57
	KEY_COLON  = 58
	KEY_SEMICOLON  = 59
	KEY_LESS  = 60
	KEY_EQUAL  = 61
	KEY_GREATER  = 62
	KEY_QUESTION  = 63
	KEY_AT  = 64
	KEY_A  = 65
	KEY_B  = 66
	KEY_C  = 67
	KEY_D  = 68
	KEY_E  = 69
	KEY_F  = 70
	KEY_G  = 71
	KEY_H  = 72
	KEY_I  = 73
	KEY_J  = 74
	KEY_K  = 75
	KEY_L  = 76
	KEY_M  = 77
	KEY_N  = 78
	KEY_O  = 79
	KEY_P  = 80
	KEY_Q  = 81
	KEY_R  = 82
	KEY_S  = 83
	KEY_T  = 84
	KEY_U  = 85
	KEY_V  = 86
	KEY_W  = 87
	KEY_X  = 88
	KEY_Y  = 89
	KEY_Z  = 90
	KEY_BRACKETLEFT  = 91
	KEY_BACKSLASH  = 92
	KEY_BRACKETRIGHT  = 93
	KEY_ASCIICIRCUM  = 94
	KEY_UNDERSCORE  = 95
	KEY_QUOTELEFT  = 96
	KEY_BRACELEFT  = 123
	KEY_BAR  = 124
	KEY_BRACERIGHT  = 125
	KEY_ASCIITILDE  = 126
	KEY_NOBREAKSPACE  = 160
	KEY_EXCLAMDOWN  = 161
	KEY_CENT  = 162
	KEY_STERLING  = 163
	KEY_CURRENCY  = 164
	KEY_YEN  = 165
	KEY_BROKENBAR  = 166
	KEY_SECTION  = 167
	KEY_DIAERESIS  = 168
	KEY_COPYRIGHT  = 169
	KEY_ORDFEMININE  = 170
	KEY_GUILLEMOTLEFT  = 171
	KEY_NOTSIGN  = 172
	KEY_HYPHEN  = 173
	KEY_REGISTERED  = 174
	KEY_MACRON  = 175
	KEY_DEGREE  = 176
	KEY_PLUSMINUS  = 177
	KEY_TWOSUPERIOR  = 178
	KEY_THREESUPERIOR  = 179
	KEY_ACUTE  = 180
	KEY_MU  = 181
	KEY_PARAGRAPH  = 182
	KEY_PERIODCENTERED  = 183
	KEY_CEDILLA  = 184
	KEY_ONESUPERIOR  = 185
	KEY_MASCULINE  = 186
	KEY_GUILLEMOTRIGHT  = 187
	KEY_ONEQUARTER  = 188
	KEY_ONEHALF  = 189
	KEY_THREEQUARTERS  = 190
	KEY_QUESTIONDOWN  = 191
	KEY_AGRAVE  = 192
	KEY_AACUTE  = 193
	KEY_ACIRCUMFLEX  = 194
	KEY_ATILDE  = 195
	KEY_ADIAERESIS  = 196
	KEY_ARING  = 197
	KEY_AE  = 198
	KEY_CCEDILLA  = 199
	KEY_EGRAVE  = 200
	KEY_EACUTE  = 201
	KEY_ECIRCUMFLEX  = 202
	KEY_EDIAERESIS  = 203
	KEY_IGRAVE  = 204
	KEY_IACUTE  = 205
	KEY_ICIRCUMFLEX  = 206
	KEY_IDIAERESIS  = 207
	KEY_ETH  = 208
	KEY_NTILDE  = 209
	KEY_OGRAVE  = 210
	KEY_OACUTE  = 211
	KEY_OCIRCUMFLEX  = 212
	KEY_OTILDE  = 213
	KEY_ODIAERESIS  = 214
	KEY_MULTIPLY  = 215
	KEY_OOBLIQUE  = 216
	KEY_UGRAVE  = 217
	KEY_UACUTE  = 218
	KEY_UCIRCUMFLEX  = 219
	KEY_UDIAERESIS  = 220
	KEY_YACUTE  = 221
	KEY_THORN  = 222
	KEY_SSHARP  = 223
	KEY_DIVISION  = 247
	KEY_YDIAERESIS  = 255
	)
type KeyModifierMask int
const (
	KEY_CODE_MASK KeyModifierMask = 8388607
	KEY_MODIFIER_MASK  = 532676608
	KEY_MASK_CMD_OR_CTRL  = 16777216
	KEY_MASK_SHIFT  = 33554432
	KEY_MASK_ALT  = 67108864
	KEY_MASK_META  = 134217728
	KEY_MASK_CTRL  = 268435456
	KEY_MASK_KPAD  = 536870912
	KEY_MASK_GROUP_SWITCH  = 1073741824
	)
type MouseButton int
const (
	MOUSE_BUTTON_NONE MouseButton = 0
	MOUSE_BUTTON_LEFT  = 1
	MOUSE_BUTTON_RIGHT  = 2
	MOUSE_BUTTON_MIDDLE  = 3
	MOUSE_BUTTON_WHEEL_UP  = 4
	MOUSE_BUTTON_WHEEL_DOWN  = 5
	MOUSE_BUTTON_WHEEL_LEFT  = 6
	MOUSE_BUTTON_WHEEL_RIGHT  = 7
	MOUSE_BUTTON_XBUTTON1  = 8
	MOUSE_BUTTON_XBUTTON2  = 9
	MOUSE_BUTTON_MASK_LEFT  = 1
	MOUSE_BUTTON_MASK_RIGHT  = 2
	MOUSE_BUTTON_MASK_MIDDLE  = 4
	MOUSE_BUTTON_MASK_XBUTTON1  = 128
	MOUSE_BUTTON_MASK_XBUTTON2  = 256
	)
type JoyButton int
const (
	JOY_BUTTON_INVALID JoyButton = -1
	JOY_BUTTON_A  = 0
	JOY_BUTTON_B  = 1
	JOY_BUTTON_X  = 2
	JOY_BUTTON_Y  = 3
	JOY_BUTTON_BACK  = 4
	JOY_BUTTON_GUIDE  = 5
	JOY_BUTTON_START  = 6
	JOY_BUTTON_LEFT_STICK  = 7
	JOY_BUTTON_RIGHT_STICK  = 8
	JOY_BUTTON_LEFT_SHOULDER  = 9
	JOY_BUTTON_RIGHT_SHOULDER  = 10
	JOY_BUTTON_DPAD_UP  = 11
	JOY_BUTTON_DPAD_DOWN  = 12
	JOY_BUTTON_DPAD_LEFT  = 13
	JOY_BUTTON_DPAD_RIGHT  = 14
	JOY_BUTTON_MISC1  = 15
	JOY_BUTTON_PADDLE1  = 16
	JOY_BUTTON_PADDLE2  = 17
	JOY_BUTTON_PADDLE3  = 18
	JOY_BUTTON_PADDLE4  = 19
	JOY_BUTTON_TOUCHPAD  = 20
	JOY_BUTTON_SDL_MAX  = 21
	JOY_BUTTON_MAX  = 128
	)
type JoyAxis int
const (
	JOY_AXIS_INVALID JoyAxis = -1
	JOY_AXIS_LEFT_X  = 0
	JOY_AXIS_LEFT_Y  = 1
	JOY_AXIS_RIGHT_X  = 2
	JOY_AXIS_RIGHT_Y  = 3
	JOY_AXIS_TRIGGER_LEFT  = 4
	JOY_AXIS_TRIGGER_RIGHT  = 5
	JOY_AXIS_SDL_MAX  = 6
	JOY_AXIS_MAX  = 10
	)
type MIDIMessage int
const (
	MIDI_MESSAGE_NONE MIDIMessage = 0
	MIDI_MESSAGE_NOTE_OFF  = 8
	MIDI_MESSAGE_NOTE_ON  = 9
	MIDI_MESSAGE_AFTERTOUCH  = 10
	MIDI_MESSAGE_CONTROL_CHANGE  = 11
	MIDI_MESSAGE_PROGRAM_CHANGE  = 12
	MIDI_MESSAGE_CHANNEL_PRESSURE  = 13
	MIDI_MESSAGE_PITCH_BEND  = 14
	MIDI_MESSAGE_SYSTEM_EXCLUSIVE  = 240
	MIDI_MESSAGE_QUARTER_FRAME  = 241
	MIDI_MESSAGE_SONG_POSITION_POINTER  = 242
	MIDI_MESSAGE_SONG_SELECT  = 243
	MIDI_MESSAGE_TUNE_REQUEST  = 246
	MIDI_MESSAGE_TIMING_CLOCK  = 248
	MIDI_MESSAGE_START  = 250
	MIDI_MESSAGE_CONTINUE  = 251
	MIDI_MESSAGE_STOP  = 252
	MIDI_MESSAGE_ACTIVE_SENSING  = 254
	MIDI_MESSAGE_SYSTEM_RESET  = 255
	)
type Error int
const (
	OK Error = 0
	FAILED  = 1
	ERR_UNAVAILABLE  = 2
	ERR_UNCONFIGURED  = 3
	ERR_UNAUTHORIZED  = 4
	ERR_PARAMETER_RANGE_ERROR  = 5
	ERR_OUT_OF_MEMORY  = 6
	ERR_FILE_NOT_FOUND  = 7
	ERR_FILE_BAD_DRIVE  = 8
	ERR_FILE_BAD_PATH  = 9
	ERR_FILE_NO_PERMISSION  = 10
	ERR_FILE_ALREADY_IN_USE  = 11
	ERR_FILE_CANT_OPEN  = 12
	ERR_FILE_CANT_WRITE  = 13
	ERR_FILE_CANT_READ  = 14
	ERR_FILE_UNRECOGNIZED  = 15
	ERR_FILE_CORRUPT  = 16
	ERR_FILE_MISSING_DEPENDENCIES  = 17
	ERR_FILE_EOF  = 18
	ERR_CANT_OPEN  = 19
	ERR_CANT_CREATE  = 20
	ERR_QUERY_FAILED  = 21
	ERR_ALREADY_IN_USE  = 22
	ERR_LOCKED  = 23
	ERR_TIMEOUT  = 24
	ERR_CANT_CONNECT  = 25
	ERR_CANT_RESOLVE  = 26
	ERR_CONNECTION_ERROR  = 27
	ERR_CANT_ACQUIRE_RESOURCE  = 28
	ERR_CANT_FORK  = 29
	ERR_INVALID_DATA  = 30
	ERR_INVALID_PARAMETER  = 31
	ERR_ALREADY_EXISTS  = 32
	ERR_DOES_NOT_EXIST  = 33
	ERR_DATABASE_CANT_READ  = 34
	ERR_DATABASE_CANT_WRITE  = 35
	ERR_COMPILATION_FAILED  = 36
	ERR_METHOD_NOT_FOUND  = 37
	ERR_LINK_FAILED  = 38
	ERR_SCRIPT_FAILED  = 39
	ERR_CYCLIC_LINK  = 40
	ERR_INVALID_DECLARATION  = 41
	ERR_DUPLICATE_SYMBOL  = 42
	ERR_PARSE_ERROR  = 43
	ERR_BUSY  = 44
	ERR_SKIP  = 45
	ERR_HELP  = 46
	ERR_BUG  = 47
	ERR_PRINTER_ON_FIRE  = 48
	)
type PropertyHint int
const (
	PROPERTY_HINT_NONE PropertyHint = 0
	PROPERTY_HINT_RANGE  = 1
	PROPERTY_HINT_ENUM  = 2
	PROPERTY_HINT_ENUM_SUGGESTION  = 3
	PROPERTY_HINT_EXP_EASING  = 4
	PROPERTY_HINT_LINK  = 5
	PROPERTY_HINT_FLAGS  = 6
	PROPERTY_HINT_LAYERS_2D_RENDER  = 7
	PROPERTY_HINT_LAYERS_2D_PHYSICS  = 8
	PROPERTY_HINT_LAYERS_2D_NAVIGATION  = 9
	PROPERTY_HINT_LAYERS_3D_RENDER  = 10
	PROPERTY_HINT_LAYERS_3D_PHYSICS  = 11
	PROPERTY_HINT_LAYERS_3D_NAVIGATION  = 12
	PROPERTY_HINT_FILE  = 13
	PROPERTY_HINT_DIR  = 14
	PROPERTY_HINT_GLOBAL_FILE  = 15
	PROPERTY_HINT_GLOBAL_DIR  = 16
	PROPERTY_HINT_RESOURCE_TYPE  = 17
	PROPERTY_HINT_MULTILINE_TEXT  = 18
	PROPERTY_HINT_EXPRESSION  = 19
	PROPERTY_HINT_PLACEHOLDER_TEXT  = 20
	PROPERTY_HINT_COLOR_NO_ALPHA  = 21
	PROPERTY_HINT_IMAGE_COMPRESS_LOSSY  = 22
	PROPERTY_HINT_IMAGE_COMPRESS_LOSSLESS  = 23
	PROPERTY_HINT_OBJECT_ID  = 24
	PROPERTY_HINT_TYPE_STRING  = 25
	PROPERTY_HINT_NODE_PATH_TO_EDITED_NODE  = 26
	PROPERTY_HINT_METHOD_OF_VARIANT_TYPE  = 27
	PROPERTY_HINT_METHOD_OF_BASE_TYPE  = 28
	PROPERTY_HINT_METHOD_OF_INSTANCE  = 29
	PROPERTY_HINT_METHOD_OF_SCRIPT  = 30
	PROPERTY_HINT_PROPERTY_OF_VARIANT_TYPE  = 31
	PROPERTY_HINT_PROPERTY_OF_BASE_TYPE  = 32
	PROPERTY_HINT_PROPERTY_OF_INSTANCE  = 33
	PROPERTY_HINT_PROPERTY_OF_SCRIPT  = 34
	PROPERTY_HINT_OBJECT_TOO_BIG  = 35
	PROPERTY_HINT_NODE_PATH_VALID_TYPES  = 36
	PROPERTY_HINT_SAVE_FILE  = 37
	PROPERTY_HINT_GLOBAL_SAVE_FILE  = 38
	PROPERTY_HINT_INT_IS_OBJECTID  = 39
	PROPERTY_HINT_INT_IS_POINTER  = 41
	PROPERTY_HINT_ARRAY_TYPE  = 40
	PROPERTY_HINT_LOCALE_ID  = 42
	PROPERTY_HINT_LOCALIZABLE_STRING  = 43
	PROPERTY_HINT_NODE_TYPE  = 44
	PROPERTY_HINT_HIDE_QUATERNION_EDIT  = 45
	PROPERTY_HINT_PASSWORD  = 46
	PROPERTY_HINT_MAX  = 47
	)
type PropertyUsageFlags int
const (
	PROPERTY_USAGE_NONE PropertyUsageFlags = 0
	PROPERTY_USAGE_STORAGE  = 2
	PROPERTY_USAGE_EDITOR  = 4
	PROPERTY_USAGE_CHECKABLE  = 8
	PROPERTY_USAGE_CHECKED  = 16
	PROPERTY_USAGE_INTERNATIONALIZED  = 32
	PROPERTY_USAGE_GROUP  = 64
	PROPERTY_USAGE_CATEGORY  = 128
	PROPERTY_USAGE_SUBGROUP  = 256
	PROPERTY_USAGE_CLASS_IS_BITFIELD  = 512
	PROPERTY_USAGE_NO_INSTANCE_STATE  = 1024
	PROPERTY_USAGE_RESTART_IF_CHANGED  = 2048
	PROPERTY_USAGE_SCRIPT_VARIABLE  = 4096
	PROPERTY_USAGE_STORE_IF_NULL  = 8192
	PROPERTY_USAGE_ANIMATE_AS_TRIGGER  = 16384
	PROPERTY_USAGE_UPDATE_ALL_IF_MODIFIED  = 32768
	PROPERTY_USAGE_SCRIPT_DEFAULT_VALUE  = 65536
	PROPERTY_USAGE_CLASS_IS_ENUM  = 131072
	PROPERTY_USAGE_NIL_IS_VARIANT  = 262144
	PROPERTY_USAGE_INTERNAL  = 524288
	PROPERTY_USAGE_DO_NOT_SHARE_ON_DUPLICATE  = 1048576
	PROPERTY_USAGE_HIGH_END_GFX  = 2097152
	PROPERTY_USAGE_NODE_PATH_FROM_SCENE_ROOT  = 4194304
	PROPERTY_USAGE_RESOURCE_NOT_PERSISTENT  = 8388608
	PROPERTY_USAGE_KEYING_INCREMENTS  = 16777216
	PROPERTY_USAGE_DEFERRED_SET_RESOURCE  = 33554432
	PROPERTY_USAGE_EDITOR_INSTANTIATE_OBJECT  = 67108864
	PROPERTY_USAGE_EDITOR_BASIC_SETTING  = 134217728
	PROPERTY_USAGE_READ_ONLY  = 268435456
	PROPERTY_USAGE_ARRAY  = 536870912
	PROPERTY_USAGE_DEFAULT  = 6
	PROPERTY_USAGE_DEFAULT_INTL  = 38
	PROPERTY_USAGE_NO_EDITOR  = 2
	)
type MethodFlags int
const (
	METHOD_FLAG_NORMAL MethodFlags = 1
	METHOD_FLAG_EDITOR  = 2
	METHOD_FLAG_CONST  = 4
	METHOD_FLAG_VIRTUAL  = 8
	METHOD_FLAG_VARARG  = 16
	METHOD_FLAG_STATIC  = 32
	METHOD_FLAG_OBJECT_CORE  = 64
	METHOD_FLAGS_DEFAULT  = 1
	)
type VariantType int
const (
	TYPE_NIL VariantType = 0
	TYPE_BOOL  = 1
	TYPE_INT  = 2
	TYPE_FLOAT  = 3
	TYPE_STRING  = 4
	TYPE_VECTOR2  = 5
	TYPE_VECTOR2I  = 6
	TYPE_RECT2  = 7
	TYPE_RECT2I  = 8
	TYPE_VECTOR3  = 9
	TYPE_VECTOR3I  = 10
	TYPE_TRANSFORM2D  = 11
	TYPE_VECTOR4  = 12
	TYPE_VECTOR4I  = 13
	TYPE_PLANE  = 14
	TYPE_QUATERNION  = 15
	TYPE_AABB  = 16
	TYPE_BASIS  = 17
	TYPE_TRANSFORM3D  = 18
	TYPE_PROJECTION  = 19
	TYPE_COLOR  = 20
	TYPE_STRING_NAME  = 21
	TYPE_NODE_PATH  = 22
	TYPE_RID  = 23
	TYPE_OBJECT  = 24
	TYPE_CALLABLE  = 25
	TYPE_SIGNAL  = 26
	TYPE_DICTIONARY  = 27
	TYPE_ARRAY  = 28
	TYPE_PACKED_BYTE_ARRAY  = 29
	TYPE_PACKED_INT32_ARRAY  = 30
	TYPE_PACKED_INT64_ARRAY  = 31
	TYPE_PACKED_FLOAT32_ARRAY  = 32
	TYPE_PACKED_FLOAT64_ARRAY  = 33
	TYPE_PACKED_STRING_ARRAY  = 34
	TYPE_PACKED_VECTOR2_ARRAY  = 35
	TYPE_PACKED_VECTOR3_ARRAY  = 36
	TYPE_PACKED_COLOR_ARRAY  = 37
	TYPE_MAX  = 38
	)
type VariantOperator int
const (
	OP_EQUAL VariantOperator = 0
	OP_NOT_EQUAL  = 1
	OP_LESS  = 2
	OP_LESS_EQUAL  = 3
	OP_GREATER  = 4
	OP_GREATER_EQUAL  = 5
	OP_ADD  = 6
	OP_SUBTRACT  = 7
	OP_MULTIPLY  = 8
	OP_DIVIDE  = 9
	OP_NEGATE  = 10
	OP_POSITIVE  = 11
	OP_MODULE  = 12
	OP_POWER  = 13
	OP_SHIFT_LEFT  = 14
	OP_SHIFT_RIGHT  = 15
	OP_BIT_AND  = 16
	OP_BIT_OR  = 17
	OP_BIT_XOR  = 18
	OP_BIT_NEGATE  = 19
	OP_AND  = 20
	OP_OR  = 21
	OP_XOR  = 22
	OP_NOT  = 23
	OP_IN  = 24
	OP_MAX  = 25
	)

