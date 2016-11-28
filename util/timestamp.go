package util

import "time"

func NumberToTime(num int64) time.Time {
	return time.Unix(0, num*1000000)
}

func TimeToNumber(time time.Time) int64 {
	return time.UnixNano() / 1000000
}
