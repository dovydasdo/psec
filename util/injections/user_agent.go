package util

type UserAgentInfo struct {
	UserAgent      string
	AcceptLanguage string
	Platform       string
	Metadata       UAMetadata
}

type UAMetadata struct {
	Brands          []UABrands
	FullVersionList []UABrands
	FullVersion     string
	Platform        string
	PlatformVersion string
	Architecture    string
	Model           string
	Mobile          bool
	Bitness         string
	WOW64           bool
}

type UABrands struct {
	Brand   string
	Version string
}

func GetStaticUAInfo() UserAgentInfo {
	// Todo: fetch from https://github.com/ulixee/unblocked-emulator-data
	return UserAgentInfo{
		UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36",
		AcceptLanguage: "en-US",
		Platform:       "Win32",
		Metadata: UAMetadata{
			Brands: []UABrands{
				{
					Brand:   "Chromium",
					Version: "118",
				},
				{
					Brand:   "Google Chrome",
					Version: "118",
				},
				{
					Brand:   "Not=A?Brand",
					Version: "99",
				},
			},
			FullVersionList: []UABrands{
				{
					Brand:   "Chromium",
					Version: "118.0.5993.71",
				},
				{
					Brand:   "Google Chrome",
					Version: "118.0.5993.71",
				},
				{
					Brand:   "Not=A?Brand",
					Version: "99.0.0.0",
				},
			},
			FullVersion:     "118.0.5993.71",
			Platform:        "Windows",
			PlatformVersion: "6.0.0",
			Architecture:    "x86",
			Model:           "",
			Mobile:          false,
			Bitness:         "64",
			WOW64:           false,
		},
	}
}
