package types

import "time"

type Result struct {
	TotalBitcoin         float64
	Date                 time.Time
	TotalBitcoinOverride float64
}
