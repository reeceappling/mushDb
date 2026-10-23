package unix

import "time"

// Time (UnixMilli) is equivalent to milliseconds since the Unix Epoch
type Time int64

func (t Time) GoTime() time.Time {
	sec := t / 1000
	return time.Unix(int64(sec), 0)
}

// TimeFor grabs the unix.Time given a go stdlib time
func TimeFor(t time.Time) Time {
	return Time(t.UnixMilli())
}

// TimeForNow grabs the unix.Time for the current go stdlib time
func TimeForNow() Time {
	return TimeFor(time.Now())
}
