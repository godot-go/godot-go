package pkg

import (
	. "github.com/godot-go/godot-go/pkg/builtin"
)

// Synchronous variant matrix audit (openspec change
// support-variant-reflect-completeness, task 6.3).
//
// Each TestAudit* method is a pure echo for exactly one GDExtension variant
// type. The GDScript side (test/demo/main.gd: test_variant_matrix_audit)
// drives every echo twice -- a direct call (ptrcall fast path) and
// example.call() (varcall) -- and requires the value back equal. Together
// with the argument and return metadata the binder derives from these
// signatures, the matrix pins: mapper acceptance, varcall decode, ptrcall
// decode, varcall return encode, ptrcall return encode, for every variant
// type at one call site per cell.
//
// The ten packed array rows reuse the slice echoes bound by
// support-native-slice-marshalling (test_slice_echo_*); they are exercised in
// the GDScript matrix, not duplicated here. NIL is audited through the
// Variant echo (a null argument is a NIL variant), VARIANT_MAX through the
// same echo with typed values.

func (e *Example) TestAuditBool(v bool) bool { return v }

func (e *Example) TestAuditInt(v int64) int64 { return v }

func (e *Example) TestAuditFloat(v float64) float64 { return v }

func (e *Example) TestAuditString(v string) string { return v }

func (e *Example) TestAuditVector2(v Vector2) Vector2 { return v }

func (e *Example) TestAuditVector2i(v Vector2i) Vector2i { return v }

func (e *Example) TestAuditRect2(v Rect2) Rect2 { return v }

func (e *Example) TestAuditRect2i(v Rect2i) Rect2i { return v }

func (e *Example) TestAuditVector3(v Vector3) Vector3 { return v }

func (e *Example) TestAuditVector3i(v Vector3i) Vector3i { return v }

func (e *Example) TestAuditTransform2D(v Transform2D) Transform2D { return v }

func (e *Example) TestAuditVector4(v Vector4) Vector4 { return v }

func (e *Example) TestAuditVector4i(v Vector4i) Vector4i { return v }

func (e *Example) TestAuditPlane(v Plane) Plane { return v }

func (e *Example) TestAuditQuaternion(v Quaternion) Quaternion { return v }

func (e *Example) TestAuditAABB(v AABB) AABB { return v }

func (e *Example) TestAuditBasis(v Basis) Basis { return v }

func (e *Example) TestAuditTransform3D(v Transform3D) Transform3D { return v }

func (e *Example) TestAuditProjection(v Projection) Projection { return v }

func (e *Example) TestAuditColor(v Color) Color { return v }

func (e *Example) TestAuditStringName(v StringName) StringName { return v }

func (e *Example) TestAuditNodePath(v NodePath) NodePath { return v }

func (e *Example) TestAuditRID(v RID) RID { return v }

// TestAuditObject echoes an Object argument. GDScript compares instance ids,
// which pins that both paths decoded the caller's instance rather than a
// substitute.
func (e *Example) TestAuditObject(v Object) Object { return v }

func (e *Example) TestAuditCallable(v Callable) Callable { return v }

func (e *Example) TestAuditSignal(v Signal) Signal { return v }

func (e *Example) TestAuditDictionary(v Dictionary) Dictionary { return v }

func (e *Example) TestAuditArray(v Array) Array { return v }

// TestAuditVariant echoes the Variant argument untouched: null exercises the
// NIL row and any typed value exercises VARIANT_MAX on both paths.
func (e *Example) TestAuditVariant(v Variant) Variant { return v }
