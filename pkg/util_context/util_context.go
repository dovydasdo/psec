package utilcontext

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type UtilInterface interface {
	ParseTime(timeStr string) (time.Time, error)
}

type UtilContext struct {
}

func New() *UtilContext {
	return &UtilContext{}
}

func (u *UtilContext) ParseTime(timeStr string) (time.Time, error) {
	nowTime := time.Now() //time.Parse("2006-01-02", timeStr)

	if strings.Contains(timeStr, "min") {
		return nowTime, nil
	}

	if isRelative(timeStr) {
		return parseRelativeDate(timeStr)
	} else {
		dateParts := strings.Split(timeStr, " ")
		mounthIdx, err := getMountIndex(dateParts[0])
		if err != nil {
			return nowTime, err
		}

		year := strconv.FormatInt(int64(nowTime.Year()), 10)
		dayToParse := dateParts[1]
		if len(dayToParse) < 2 {
			dayToParse = "0" + dayToParse
		}

		dayToParse = strings.ReplaceAll(dayToParse, ",", "")
		return time.Parse("2006-01-02", fmt.Sprintf("%s-%s-%s", year, mounthIdx, dayToParse))
	}
}

func parseRelativeDate(date string) (time.Time, error) {
	pattern := `(\d+)`

	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(date)

	if len(match) < 1 {
		return time.Now(), errors.New("regex failed to match time offset from string")
	}

	offset, err := strconv.Atoi(match[0])
	if err != nil {
		return time.Now(), errors.New("failed to parse matched to int")
	}

	return time.Now().Add(time.Duration(offset) * 24 * time.Hour), nil
}

func isRelative(date string) bool {
	return strings.Contains(date, "prieš")
}

func getMountIndex(mounth string) (string, error) {
	switch mounth {
	case "sausio":
		return "01", nil
	case "vasario":
		return "02", nil
	case "kovo":
		return "03", nil
	case "balandžio":
		return "04", nil
	case "gegužės":
		return "05", nil
	case "birželio":
		return "06", nil
	case "liepos":
		return "07", nil
	case "rugpjūčio":
		return "08", nil
	case "rugsėjo":
		return "09", nil
	case "spalio":
		return "10", nil
	case "lapkričio":
		return "11", nil
	case "gruodžio":
		return "12", nil
	default:
		return "-1", errors.New("mount not found")
	}
}
