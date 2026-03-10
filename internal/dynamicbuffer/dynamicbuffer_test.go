package dynamicbuffer_test

import (
	"testing"

	"github.com/gamefabric/terraform-provider-gamefabric/internal/dynamicbuffer"
	"github.com/stretchr/testify/assert"
)

func TestDynamicMaxBufferThreshold(t *testing.T) {
	tests := []struct {
		maxBufferUtilization int32
		wanted               int32
	}{
		{maxBufferUtilization: 85, wanted: 100},
		// --
		{maxBufferUtilization: 80, wanted: 100},
		{maxBufferUtilization: 77, wanted: 104},
		{maxBufferUtilization: 75, wanted: 107},
		{maxBufferUtilization: 72, wanted: 111},
		{maxBufferUtilization: 70, wanted: 114},
		{maxBufferUtilization: 68, wanted: 117},
		{maxBufferUtilization: 65, wanted: 121},
		{maxBufferUtilization: 63, wanted: 124},
		{maxBufferUtilization: 60, wanted: 129},
		{maxBufferUtilization: 58, wanted: 132},
		{maxBufferUtilization: 55, wanted: 136},
		{maxBufferUtilization: 53, wanted: 139},
		{maxBufferUtilization: 50, wanted: 143},
		{maxBufferUtilization: 48, wanted: 166},
		{maxBufferUtilization: 45, wanted: 200},
		{maxBufferUtilization: 43, wanted: 223},
		{maxBufferUtilization: 40, wanted: 257},
		{maxBufferUtilization: 38, wanted: 260},
		{maxBufferUtilization: 35, wanted: 264},
		{maxBufferUtilization: 33, wanted: 267},
		{maxBufferUtilization: 30, wanted: 271},
		{maxBufferUtilization: 28, wanted: 274},
		{maxBufferUtilization: 25, wanted: 279},
		{maxBufferUtilization: 23, wanted: 282},
		{maxBufferUtilization: 20, wanted: 286},
		{maxBufferUtilization: 18, wanted: 289},
		{maxBufferUtilization: 15, wanted: 293},
		{maxBufferUtilization: 13, wanted: 296},
		{maxBufferUtilization: 10, wanted: 300},
		// --
		{maxBufferUtilization: 8, wanted: 300},
	}

	for _, test := range tests {
		got := dynamicbuffer.DynamicMaxBufferThreshold(test.maxBufferUtilization)
		assert.Equalf(t,
			test.wanted,
			got,
			"Expected dynamic_max_buffer_threshold=%d, got=%d for max_buffer_utilization=%d",
			test.wanted,
			got,
			test.maxBufferUtilization,
		)
	}
}

func TestDynamicMinBufferThreshold(t *testing.T) {
	tests := []struct {
		maxBufferUtilization int32
		wanted               int32
	}{
		{maxBufferUtilization: 85, wanted: 50},
		// --
		{maxBufferUtilization: 80, wanted: 50},
		{maxBufferUtilization: 77, wanted: 48},
		{maxBufferUtilization: 75, wanted: 47},
		{maxBufferUtilization: 72, wanted: 45},
		{maxBufferUtilization: 70, wanted: 44},
		{maxBufferUtilization: 68, wanted: 43},
		{maxBufferUtilization: 65, wanted: 41},
		{maxBufferUtilization: 63, wanted: 40},
		{maxBufferUtilization: 60, wanted: 39},
		{maxBufferUtilization: 58, wanted: 38},
		{maxBufferUtilization: 55, wanted: 36},
		{maxBufferUtilization: 53, wanted: 35},
		{maxBufferUtilization: 50, wanted: 33},
		{maxBufferUtilization: 48, wanted: 32},
		{maxBufferUtilization: 45, wanted: 30},
		{maxBufferUtilization: 43, wanted: 29},
		{maxBufferUtilization: 40, wanted: 27},
		{maxBufferUtilization: 38, wanted: 26},
		{maxBufferUtilization: 35, wanted: 24},
		{maxBufferUtilization: 33, wanted: 23},
		{maxBufferUtilization: 30, wanted: 21},
		{maxBufferUtilization: 28, wanted: 20},
		{maxBufferUtilization: 25, wanted: 19},
		{maxBufferUtilization: 23, wanted: 18},
		{maxBufferUtilization: 20, wanted: 16},
		{maxBufferUtilization: 18, wanted: 15},
		{maxBufferUtilization: 15, wanted: 13},
		{maxBufferUtilization: 13, wanted: 12},
		{maxBufferUtilization: 10, wanted: 10},
		// --
		{maxBufferUtilization: 5, wanted: 10},
	}

	for _, test := range tests {
		got := dynamicbuffer.DynamicMinBufferThreshold(test.maxBufferUtilization)
		assert.Equalf(t,
			test.wanted,
			got,
			"Expected dynamic_min_buffer_threshold=%d, got=%d for max_buffer_utilization=%d",
			test.wanted,
			got,
			test.maxBufferUtilization,
		)
	}
}
