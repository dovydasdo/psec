package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
)

// Consider options for other browsers if bidi support is added
// TODO: Dont rely on hardcoded links

type UAMetadata struct {
	JsHighEntropyHints JsHighEntropyHints `json:"jsHighEntropyHints"`
}

type JsHighEntropyHints struct {
	Architecture    string            `json:"architecture"`
	Bitness         string            `json:"bitness"`
	Brands          []Brand           `json:"brands"`
	FullVersionList []FullVersionList `json:"fullVersionList"`
	Mobile          bool              `json:"mobile"`
	Model           string            `json:"model"`
	Platform        string            `json:"platform"`
	PlatformVersion string            `json:"platformVersion"`
	UaFullVersion   string            `json:"uaFullVersion"`
	Wow64           bool              `json:"wow64"`
}

type NavigatorInfo struct {
	Navigator struct {
		UserAgent struct {
			Value string `json:"_$value"`
		} `json:"userAgent"`
		Platform struct {
			Value string `json:"_$value"`
		} `json:"platform"`
	} `json:"navigator"`
}

type Brand struct {
	Brand   string `json:"brand"`
	Version string `json:"version"`
}

type FullVersionList struct {
	Brand   string `json:"brand"`
	Version string `json:"version"`
}

type UserAgentInfo struct {
	UserAgent      string
	AcceptLanguage string
	Platform       string
	Metadata       UAMetadata
}

type VersionInfo []struct {
	Version int `json:"majorVersion"`
}

func GetClientHints(version int) (UAMetadata, error) {
	// have some backup info providers
	return fetch[UAMetadata](fmt.Sprintf("https://raw.githubusercontent.com/ulixee/unblocked-emulator-data/main/as-chrome-%v-0/as-windows-10/user-agent-hints.json", version))
	// todo: some validation?
}

func GetNavigator(version int) (NavigatorInfo, error) {
	return fetch[NavigatorInfo](fmt.Sprintf("https://raw.githubusercontent.com/ulixee/unblocked-emulator-data/main/as-chrome-%v-0/as-windows-10/window-navigator.json", version))
	// todo: some validation?
}

func GetLatestInfo(lang string) (UserAgentInfo, error) {
	data := UserAgentInfo{}
	latestInfo, err := fetch[VersionInfo]("https://raw.githubusercontent.com/ulixee/unblocked-emulator-data/main/browserEngineOptions.json")

	if len(latestInfo) < 1 {
		return data, errors.New("no info about versions found")
	}

	latest := math.MinInt
	for _, i := range latestInfo {
		if latest < int(i.Version) {
			latest = int(i.Version)
		}
	}

	ch, err := GetClientHints(latest)
	if err != nil {
		return data, err
	}

	nav, err := GetNavigator(latest)
	if err != nil {
		return data, err
	}

	data.UserAgent = nav.Navigator.UserAgent.Value
	data.Platform = nav.Navigator.Platform.Value
	data.Metadata = ch
	data.AcceptLanguage = lang

	return data, err // todo: some validation?
}

func fetch[T any](url string) (T, error) {
	data := new(T)
	resp, err := http.Get(url)
	if err != nil {
		return *data, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return *data, err
	}

	err = json.Unmarshal(body, data)

	return *data, err
}
