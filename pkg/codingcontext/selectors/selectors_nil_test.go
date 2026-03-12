package selectors

import "testing"

// TestSelectors_NilString verifies that calling String() on a nil (zero-value)
// Selectors does not panic and returns the expected "{}" representation.
func TestSelectors_NilString(t *testing.T) {
	t.Parallel()

	var s Selectors

	got := s.String()

	if got != "{}" {
		t.Errorf("String() on nil Selectors = %q, want \"{}\"", got)
	}
}

// TestSelectors_GetValue_NilReceiver verifies that GetValue on a nil Selectors
// safely returns false instead of panicking.
func TestSelectors_GetValue_NilReceiver(t *testing.T) {
	t.Parallel()

	var s Selectors

	if s.GetValue("env", "production") {
		t.Error("GetValue() on nil Selectors should return false, got true")
	}
}

// TestSelectors_GetValue_MissingKey verifies that GetValue returns false when
// the key does not exist in a non-nil Selectors map.
func TestSelectors_GetValue_MissingKey(t *testing.T) {
	t.Parallel()

	s := make(Selectors)
	s.SetValue("env", "production")

	if s.GetValue("language", "go") {
		t.Error("GetValue() for missing key should return false, got true")
	}
}

// TestSelectors_GetValue_MissingValue verifies that GetValue returns false when
// the key exists but the specific value is absent.
func TestSelectors_GetValue_MissingValue(t *testing.T) {
	t.Parallel()

	s := make(Selectors)
	s.SetValue("env", "production")

	if s.GetValue("env", "development") {
		t.Error("GetValue() for present key but absent value should return false")
	}
}

// TestSelectors_SetValue_NilReceiver verifies that SetValue on a nil Selectors
// auto-initializes the map and stores the value correctly.
func TestSelectors_SetValue_NilReceiver(t *testing.T) {
	t.Parallel()

	var s Selectors
	s.SetValue("env", "production")

	if !s.GetValue("env", "production") {
		t.Error("SetValue() on nil Selectors should auto-initialize and store the value")
	}
}
