package usecase

import "time"

func timeNowAddHours(hours int) int64 {
	return time.Now().Add(time.Duration(hours) * time.Hour).Unix()
}
