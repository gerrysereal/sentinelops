package tests

import "testing"

func TestSentinelOpsPlaceholder(t *testing.T) {
	// This placeholder keeps the CI test stage explicit while integration tests are added.
	if "sentinelops" == "" {
		t.Fatal("unexpected empty project name")
	}
}
