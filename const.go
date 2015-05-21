package ringo

const (
	SequenceMax int64 = (1 << 63) - 1

	// Following are some ring buffer size recommendations based on powers of two
	PT1Meg   = 1048576
	PT2Meg   = 2097152
	PT4Meg   = 4194304
	PT8Meg   = 8388608
	PT16Meg  = 16777216
	PT32Meg  = 33554432
	PT64Meg  = 67108864
	PT128Meg = 134217728
	PT256Meg = 268435456
)
