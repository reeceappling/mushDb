package unix

import "time"

type Time int64 // unixMilli!

func TimeFor(t time.Time) Time {
	return Time(t.UnixMilli())
}
func TimeForNow() Time {
	return TimeFor(time.Now())
}
