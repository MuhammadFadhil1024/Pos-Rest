package utils

import (
	"fmt"
	"time"
)

func GenerateTicketNumber(branchID int) string {
	now := time.Now()

	return fmt.Sprintf(
		"TRX-B%02d-%s",
		branchID,
		now.Format("20060102-150405"),
	)
}