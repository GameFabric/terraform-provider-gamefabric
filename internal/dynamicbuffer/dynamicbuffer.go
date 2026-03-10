package dynamicbuffer

import "math"

const (
	idxMbu = 0
	idxMax = 1
	idxMin = 2
)

// See https://docs.gamefabric.com/multiplayer-servers/multiplayer-services/armada-replicas-and-buffer#configuring-dynamic-buffer - "Values behind the slider".
var table = [][3]int32{
	{80, 100, 50},
	{75, 107, 47},
	{70, 114, 44},
	{65, 121, 41},
	{60, 129, 39},
	{55, 136, 36},
	{50, 143, 33},
	{45, 200, 30},
	{40, 257, 27},
	{35, 264, 24},
	{30, 271, 21},
	{25, 279, 19},
	{20, 286, 16},
	{15, 293, 13},
	{10, 300, 10},
}

// DynamicMaxBufferThreshold calculates the dynamic maximum buffer threshold based on the given maximum buffer utilization.
func DynamicMaxBufferThreshold(mbu int32) int32 {
	return dynamicBufferThreshold(mbu, idxMax)
}

// DynamicMinBufferThreshold calculates the dynamic minimum buffer threshold based on the given maximum buffer utilization.
func DynamicMinBufferThreshold(mbu int32) int32 {
	return dynamicBufferThreshold(mbu, idxMin)
}

func dynamicBufferThreshold(mbu int32, idx int) int32 {
	for i := range table {
		switch {
		case i == 0 && mbu > table[i][idxMbu]:
			// Greater than first.
			return table[0][idx]
		case i == len(table)-1 && mbu < table[i][idxMbu]:
			// Less than last.
			return table[len(table)-1][idx]
		case mbu == table[i][idxMbu]:
			// Exact match.
			return table[i][idx]
		case i < len(table)-1 && mbu < table[i][idxMbu] && mbu > table[i+1][idxMbu]:
			// Interpolate between two entries.
			return interpolate(mbu, table[i], table[i+1], idx)
		}
	}
	panic("empty dynamic buffer table") // Developer error.
}

func interpolate(mbu int32, upper, lower [3]int32, idx int) int32 {
	ratio := float64(mbu-lower[idxMbu]) / float64(upper[idxMbu]-lower[idxMbu])
	val := float64(lower[idx]) + ratio*float64(upper[idx]-lower[idx])
	return int32(math.Round(val))
}
