package util

import (
	"crypto/rand"
	"fmt"
	"time"
)

func FormatDateIt(tt time.Time) string {
	res := fmt.Sprintf("%d %s %d", tt.Day(), MonthToStringIT(tt.Month()), tt.Year())
	return res
}

func FormatDateTimeIt(tt time.Time) string {
	res := fmt.Sprintf("%d %s %d - %02d:%02d", tt.Day(), MonthToStringIT(tt.Month()), tt.Year(), tt.Local().Hour(), tt.Local().Minute())
	return res
}

func MonthToStringIT(month time.Month) string {
	switch month {
	case time.January:
		return "Gennaio"
	case time.February:
		return "Febbraio"
	case time.March:
		return "Marzo"
	case time.April:
		return "Aprile"
	case time.May:
		return "Maggio"
	case time.June:
		return "Giugno"
	case time.July:
		return "Luglio"
	case time.August:
		return "Agosto"
	case time.September:
		return "Settembre"
	case time.October:
		return "Ottobre"
	case time.November:
		return "Novembre"
	case time.December:
		return "Dicembre"
	default:
		return ""
	}
}

func PseudoUuid() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	uuid := fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	return uuid, nil
}
