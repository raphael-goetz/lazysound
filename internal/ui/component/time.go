package component

import "fmt"

func Time(mills int64) string {
	seconds := mills / 1000
	minutes := seconds / 60

	leftSec := seconds - (minutes * 60)

	if minutes < 1 {
		return fmt.Sprintf("%ds", int(seconds))
	}

	if minutes > 60 {
		hours := minutes / 60
		leftMin := minutes - (hours * 60)
		return fmt.Sprintf("%dh %dm", int(hours), int(leftMin))
	}

	return fmt.Sprintf("%dm %ds", int(minutes), int(leftSec))
}
