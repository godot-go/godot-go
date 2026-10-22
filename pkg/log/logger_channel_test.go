package log

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"

	"go.uber.org/zap"
)

// Witnesses for the diagnostic-output-channel spec. The entry threshold is
// chosen at package init, so threshold behaviour can only be observed at
// process start: the parent test re-execs this test binary in a child mode
// with a controlled LOG_LEVEL, the child emits probe entries, and the parent
// asserts on the child's captured stdout.

const childEnv = "GODOTGO_LOG_TEST_CHILD"

func TestMain(m *testing.M) {
	switch os.Getenv(childEnv) {
	case "":
		os.Exit(m.Run())
	case "threshold":
		Debug("threshold-probe-debug")
		Info("threshold-probe-info")
		Warn("threshold-probe-warn")
		Error("threshold-probe-error")
		fmt.Println("threshold-probe-complete")
		os.Exit(0)
	case "concurrent":
		emitConcurrentProbe()
		os.Exit(0)
	case "stacks":
		Debug("stacks-probe-debug")
		Warn("stacks-probe-warn")
		fmt.Println("stacks-probe-complete")
		os.Exit(0)
	case "fault":
		emitFaultProbe()
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "unknown child mode %q\n", os.Getenv(childEnv))
		os.Exit(3)
	}
}

func runLogChild(t *testing.T, mode string, level string, setLevel bool) string {
	t.Helper()
	env := make([]string, 0, len(os.Environ())+2)
	for _, kv := range os.Environ() {
		if strings.HasPrefix(kv, "LOG_LEVEL=") || strings.HasPrefix(kv, childEnv+"=") {
			continue
		}
		env = append(env, kv)
	}
	env = append(env, childEnv+"="+mode)
	if setLevel {
		env = append(env, "LOG_LEVEL="+level)
	}
	cmd := exec.Command(os.Args[0])
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("child %s (level=%q) failed: %v\noutput:\n%s", mode, level, err, out)
	}
	return string(out)
}

// 5.1: threshold selection at start — unconfigured yields the default
// (warning and above), a configured value admits quieter entries, and an
// unparseable value yields the default with start-up continuing.
func TestLogThresholdSelection(t *testing.T) {
	t.Run("unconfigured yields default", func(t *testing.T) {
		out := runLogChild(t, "threshold", "", false)
		if !strings.Contains(out, "threshold-probe-complete") {
			t.Fatal("child did not start up and finish")
		}
		for _, marker := range []string{"threshold-probe-warn", "threshold-probe-error"} {
			if !strings.Contains(out, marker) {
				t.Errorf("default threshold must admit %s; output:\n%s", marker, out)
			}
		}
		for _, marker := range []string{"threshold-probe-debug", "threshold-probe-info"} {
			if strings.Contains(out, marker) {
				t.Errorf("default threshold must suppress %s; output:\n%s", marker, out)
			}
		}
	})
	t.Run("configured admits quieter entries", func(t *testing.T) {
		out := runLogChild(t, "threshold", "info", true)
		if !strings.Contains(out, "threshold-probe-info") {
			t.Errorf("info threshold must admit info entries; output:\n%s", out)
		}
	})
	t.Run("unparseable yields default and start-up continues", func(t *testing.T) {
		out := runLogChild(t, "threshold", "not-a-level", true)
		if !strings.Contains(out, "threshold-probe-complete") {
			t.Fatalf("start-up must continue past an unparseable level; output:\n%s", out)
		}
		if !strings.Contains(out, "threshold-probe-warn") {
			t.Errorf("unparseable level must fall back to the default threshold; output:\n%s", out)
		}
		if strings.Contains(out, "threshold-probe-info") {
			t.Errorf("unparseable level must not admit info; output:\n%s", out)
		}
	})
}

// 5.2: concurrent writers never interleave — every entry arrives whole on its
// own line, and suppressed levels are absent.
func TestLogConcurrentWriters(t *testing.T) {
	const writers, perWriter = 8, 5
	out := runLogChild(t, "concurrent", "", false)
	if !strings.Contains(out, "concurrent-probe-complete") {
		t.Fatalf("child did not finish; output:\n%s", out)
	}
	found := map[string]int{}
	for _, line := range strings.Split(out, "\n") {
		inLine := 0
		for w := 0; w < writers; w++ {
			for i := 0; i < perWriter; i++ {
				marker := fmt.Sprintf("concurrent-probe-%d-%d", w, i)
				if strings.Contains(line, marker) {
					found[marker]++
					inLine++
				}
			}
		}
		if inLine > 1 {
			t.Fatalf("two entries share one line: %q", line)
		}
	}
	for w := 0; w < writers; w++ {
		for i := 0; i < perWriter; i++ {
			marker := fmt.Sprintf("concurrent-probe-%d-%d", w, i)
			if found[marker] != 1 {
				t.Errorf("entry %s arrived %d times, want exactly 1 whole occurrence", marker, found[marker])
			}
		}
	}
	if strings.Contains(out, "suppressed-debug") {
		t.Error("debug entries must be suppressed at the default threshold")
	}
}

func emitConcurrentProbe() {
	var wg sync.WaitGroup
	for w := 0; w < 8; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for i := 0; i < 5; i++ {
				Warn(fmt.Sprintf("concurrent-probe-%d-%d", w, i))
				Debug(fmt.Sprintf("suppressed-debug-%d-%d", w, i))
			}
		}(w)
	}
	wg.Wait()
	fmt.Println("concurrent-probe-complete")
}

// 5.3: a warning-level entry carries a call stack and an admitted debug entry
// carries none.
func TestLogStackPolicy(t *testing.T) {
	out := runLogChild(t, "stacks", "debug", true)
	debugIdx := strings.Index(out, "stacks-probe-debug")
	warnIdx := strings.Index(out, "stacks-probe-warn")
	endIdx := strings.Index(out, "stacks-probe-complete")
	if debugIdx < 0 || warnIdx < 0 || endIdx < 0 {
		t.Fatalf("debug threshold must admit debug and warn entries; output:\n%s", out)
	}
	if debugIdx > warnIdx {
		t.Fatal("entries must arrive in emission order")
	}
	between := out[debugIdx:warnIdx]
	if strings.Contains(between, "runtime.main") {
		t.Error("a debug entry must carry no call stack")
	}
	// zap's console encoder writes stack frames without a header, so frame
	// content (runtime.main) distinguishes a stack from the caller field.
	if !strings.Contains(out[warnIdx:endIdx], "runtime.main") {
		t.Error("a warning-level entry must carry a call stack")
	}
}

// 5.4: a binding fault records its entry before unwinding, the entry survives
// the default threshold, and it identifies the rejected construct.
func TestLogBindingFaultRecordsBeforeUnwinding(t *testing.T) {
	out := runLogChild(t, "fault", "", false)
	faultIdx := strings.Index(out, "binding-fault-probe rejected ZetaWidget")
	sentinelIdx := strings.Index(out, "fault-probe-complete")
	if faultIdx < 0 {
		t.Fatalf("the fault entry must be recorded at the default threshold; output:\n%s", out)
	}
	if sentinelIdx < 0 {
		t.Fatalf("the fault must unwind recoverably, not end the process; output:\n%s", out)
	}
	if faultIdx > sentinelIdx {
		t.Error("the fault entry must be recorded before the unwinding reaches the caller")
	}
}

func emitFaultProbe() {
	recovered := func() (r any) {
		defer func() { r = recover() }()
		Panic("binding-fault-probe rejected ZetaWidget", zap.String("construct", "ZetaWidget"))
		return nil // unreachable: Panic unwinds before this returns
	}()
	if recovered == nil {
		fmt.Fprintln(os.Stderr, "fault probe did not unwind through the panic path")
		os.Exit(2)
	}
	fmt.Println("fault-probe-complete")
}

// The fault path unwinds as a recoverable panic carrying the message, which
// is what every recovered witness in the demo relies on.
func TestPanicUnwindsWithMessage(t *testing.T) {
	var caught any
	func() {
		defer func() { caught = recover() }()
		Panic("unwind witness")
	}()
	if caught == nil {
		t.Fatal("log.Panic must unwind through the panic path")
	}
	if !strings.Contains(fmt.Sprint(caught), "unwind witness") {
		t.Errorf("the panic value must carry the message, got %v", caught)
	}
}
