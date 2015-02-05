package file

import (
	"testing"
	"time"
)

func TestGetNowDateJst(t *testing.T) {
	d := time.Date(2015, 2, 5, 20, 00, 00, 0, time.UTC)

	j := d.In(time.FixedZone("Asia/Tokyo", 9*60*60))

	r := getNowDateJst(j)
	if "20150206" != r {
		t.Fatal("Non-expected time : %v", r)
	}
}
