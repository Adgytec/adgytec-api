package controllers

import (
	"context"
	"time"
)

const mb = 1 << 20

var ctx = context.Background()

func getNow() string {
	now := time.Now()
	ist := time.FixedZone("IST", 5*60*60+30*60)
	istTime := now.In(ist)

	return istTime.Format(time.RFC3339)
}
