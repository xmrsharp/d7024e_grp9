package network

import "testing"

func UnitTestTest(t *testingT) {
	returned := 2
	expected := 2
	if returned != expected {
		t.Errorf("Got")
	}
}
