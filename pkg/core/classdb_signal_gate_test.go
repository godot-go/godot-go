package core

import "testing"

// Witnesses for the class-metadata-registration requirement "A signal name
// already carried by the class or an ancestor is rejected": the pure decision
// behind ClassDBAddSignal's duplicate gate, exercised without an engine by
// stubbing the class_has_signal probe.

func signalGateFixture(recorded ...string) *ClassInfo {
	set := make(StringSet)
	for _, name := range recorded {
		set[name] = struct{}{}
	}
	return &ClassInfo{Name: "SignalGateFixture", SignalNameSet: set}
}

func TestSignalNameCollisionRecordedName(t *testing.T) {
	ci := signalGateFixture("own_signal")
	// A recorded name collides decisively — even with no engine to consult.
	collision, decisive := signalNameCollision(ci, "own_signal", nil)
	if !collision {
		t.Fatal("a name the class recorded must collide")
	}
	if !decisive {
		t.Fatal("a decisive recorded collision must not be reported as degraded")
	}
}

func TestSignalNameCollisionInheritedName(t *testing.T) {
	ci := signalGateFixture()
	probe := func(string) (bool, bool) { return true, true }
	collision, decisive := signalNameCollision(ci, "ready", probe)
	if !collision {
		t.Fatal("a name the engine reports the class carries must collide")
	}
	if !decisive {
		t.Fatal("an engine-consulted answer must be reported as decisive")
	}
}

func TestSignalNameCollisionUnusedName(t *testing.T) {
	ci := signalGateFixture("other_signal")
	probe := func(string) (bool, bool) { return false, true }
	collision, decisive := signalNameCollision(ci, "fresh_signal", probe)
	if collision {
		t.Fatal("a name neither the class nor the engine carries must not collide")
	}
	if !decisive {
		t.Fatal("an engine-consulted answer must be reported as decisive")
	}
}

func TestSignalNameCollisionUnavailableEngine(t *testing.T) {
	ci := signalGateFixture()
	// A reporting-but-unavailable engine (nil singleton) degrades to the
	// recorded names: even an answer of true is dropped, and the caller is
	// told inheritance was not checked.
	probe := func(string) (bool, bool) { return true, false }
	collision, decisive := signalNameCollision(ci, "ready", probe)
	if collision {
		t.Fatal("an unavailable engine must not contribute a collision")
	}
	if decisive {
		t.Fatal("degraded decision must report inheritance as not checked")
	}
	// A nil probe degrades the same way.
	collision, decisive = signalNameCollision(ci, "ready", nil)
	if collision || decisive {
		t.Fatalf("nil probe: got (%v, %v), want (false, false)", collision, decisive)
	}
}
