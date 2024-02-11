package utilcontext

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	requestcontext "github.com/dovydasdo/psec/pkg/request_context"
	"golang.org/x/exp/slices"
)

func ParseTime(timeStr string, isLast bool) (time.Time, error) {
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

		yearInt := int64(nowTime.Year())

		if isLast {
			yearInt--
		}

		year := strconv.FormatInt(yearInt, 10)

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

func GetBlockList(letThrough []string, state *requestcontext.State) ([]string, error) {
	filters := make([]string, 0)

	state.NetworkEvents.Range(func(key, value interface{}) bool {
		block := true
		if ev, ok := value.(*requestcontext.NetworkEvent); ok {
			for _, pass := range letThrough {
				if strings.Contains(ev.Request.URL, pass) {
					block = false
					break
				}
			}

			if block && !slices.Contains(filters, ev.Request.URL) {
				filters = append(filters, ev.Request.URL)
			}
		}
		return true
	})

	return filters, nil
}

func GetHostBlockList(letThrough []string, state *requestcontext.State) ([]string, error) {
	filters := make([]string, 0)
	state.NetworkEvents.Range(func(key, value interface{}) bool {
		block := true
		if ev, ok := value.(*requestcontext.NetworkEvent); ok {
			eUrl, err := url.Parse(ev.Request.URL)
			if err != nil {
				log.Printf("failed to parse url: %v", ev.Request.URL)
			}

			for _, pass := range letThrough {
				if strings.Contains(eUrl.Host, pass) {
					block = false
					break
				}
			}

			if block && !slices.Contains(filters, "*"+eUrl.Host+"*") {
				filters = append(filters, "*"+eUrl.Host+"*")
			}

		}
		return true
	})

	return filters, nil
}

func SaveBlockList(list map[string][]string, path string) error {
	json, err := json.Marshal(list)
	if err != nil {
		return err
	}

	err = os.WriteFile(path, json, 0644)

	return err
}

func LoadBlockList(path string) (map[string][]string, error) {
	jsonFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		log.Printf("Error reading: %v", err)
	}
	var list map[string][]string
	err = json.Unmarshal(byteValue, &list)
	if err != nil {
		return nil, err
	}

	return list, nil
}
