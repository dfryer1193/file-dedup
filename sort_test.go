package main

import (
	"testing"
)

func TestLess(t *testing.T) {
	var tests = []struct {
		input    [2]string
		expected bool
	}{
		// Major
		{[2]string{"7.6-GA", "6.10-GA"}, false},
		{[2]string{"6.10-GA", "7.6-GA"}, true},
		// Minor
		{[2]string{"7.6-GA", "7.5-GA"}, false},
		{[2]string{"7.5-GA", "7.6-GA"}, true},
		// Release State
		{[2]string{"7.6-Alpha", "7.6-Beta"}, true},
		{[2]string{"7.6-Beta", "7.6-Snap-1"}, true},
		{[2]string{"7.6-Snap-1", "7.6-Snap-2"}, true},
		{[2]string{"7.6-Snap-2", "7.6-RC-1"}, true},
		{[2]string{"7.6-RC-1", "7.6-RC-2"}, true},
		{[2]string{"7.6-RC-2", "7.6-GA"}, true},
		{[2]string{"8.0-Alpha", "8.0-Beta"}, true},
		{[2]string{"8.0-Beta", "8.0-Beta-1.1"}, true},
		{[2]string{"8.0-Beta-1.1", "8.0-Snapshot-1"}, true},
		{[2]string{"8.0-Snapshot-1", "8.0-Snapshot-2"}, true},
		{[2]string{"8.0-Snapshot-4", "8.0.0-Snapshot-5"}, true},
		{[2]string{"8.0.0-Snapshot-5", "8.0.0-Snapshot-6"}, true},
		{[2]string{"8.0.0-Snapshot-6", "8.0.0-RC-1.0"}, true},
		{[2]string{"8.0.0-RC-1.0", "8.0.0-RC-2"}, true},
		{[2]string{"8.0.0-RC-4", "8.0.0-GA"}, true},
		// False
		{[2]string{"7.6-Beta", "7.6-Alpha"}, false},
		{[2]string{"7.6-Snap-1", "7.6-Beta"}, false},
		{[2]string{"7.6-Snap-2", "7.6-Snap-1"}, false},
		{[2]string{"7.6-RC-1", "7.6-Snap-2"}, false},
		{[2]string{"7.6-RC-2", "7.6-RC-1"}, false},
		{[2]string{"7.6-GA", "7.6-RC-2"}, false},
		{[2]string{"8.0-Beta", "8.0-Alpha"}, false},
		{[2]string{"8.0-Snapshot-1", "8.0-Beta"}, false},
		{[2]string{"8.0-Snapshot-2", "8.0-Snapshot-1"}, false},
		{[2]string{"8.0.0-Snapshot-5", "8.0-Snapshot-4"}, false},
		{[2]string{"8.0.0-Snapshot-6", "8.0.0-Snapshot-5"}, false},
		{[2]string{"8.0.0-RC-1.0", "8.0.0-Snapshot-6"}, false},
		{[2]string{"8.0.0-RC-2", "8.0.0-RC-1.0"}, false},
		{[2]string{"8.0.0-GA", "8.0.0-RC-6"}, false},
	}

	for _, test := range tests {
		if res := byRelease(test.input[:]).Less(0, 1); res != test.expected {
			t.Errorf("Test failed! %v input, %t expected. Received %t", test.input, test.expected, res)
		}
	}
}
