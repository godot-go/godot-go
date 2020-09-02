#ifndef CGDNATIVE_EXT_NET_H
#define CGDNATIVE_EXT_NET_H
/*------------------------------------------------------------------------------
//   This code was generated by template gdnative.h.tmpl.
//
//   Changes to this file may cause incorrect behavior and will be lost if
//   the code is regenerated. Any updates should be done in
//   "gdnative.h.tmpl" so they can be included in the generated
//   code.
//----------------------------------------------------------------------------*/
#include <gdnative/aabb.h>
#include <gdnative/array.h>
#include <gdnative/basis.h>
#include <gdnative/color.h>
#include <gdnative/dictionary.h>
#include <gdnative/gdnative.h>
#include <gdnative/node_path.h>
#include <gdnative/plane.h>
#include <gdnative/pool_arrays.h>
#include <gdnative/quat.h>
#include <gdnative/rect2.h>
#include <gdnative/rid.h>
#include <gdnative/string.h>
#include <gdnative/string_name.h>
#include <gdnative/transform.h>
#include <gdnative/transform2d.h>
#include <gdnative/variant.h>
#include <gdnative/vector2.h>
#include <gdnative/vector3.h>
#include <gdnative_api_struct.gen.h>

/* Go cannot call C function pointers directly, so we must generate C wrapper code to call the functions. */
/* GDNative NET 3.1 */
void go_godot_net_bind_stream_peer(godot_gdnative_ext_net_api_struct * p_api, godot_object * p_obj, const godot_net_stream_peer * p_interface);
void go_godot_net_bind_packet_peer(godot_gdnative_ext_net_api_struct * p_api, godot_object * p_obj, const godot_net_packet_peer * p_interface);
void go_godot_net_bind_multiplayer_peer(godot_gdnative_ext_net_api_struct * p_api, godot_object * p_obj, const godot_net_multiplayer_peer * p_interface);
/* GDNative NET 3.2 */
godot_error go_godot_net_set_webrtc_library(godot_gdnative_ext_net_3_2_api_struct * p_api, const godot_net_webrtc_library * p_library);
void go_godot_net_bind_webrtc_peer_connection(godot_gdnative_ext_net_3_2_api_struct * p_api, godot_object * p_obj, const godot_net_webrtc_peer_connection * p_interface);
void go_godot_net_bind_webrtc_data_channel(godot_gdnative_ext_net_3_2_api_struct * p_api, godot_object * p_obj, const godot_net_webrtc_data_channel * p_interface);
#endif
