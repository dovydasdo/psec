package util

func GetLatestUAInfo() (UserAgentInfo, error) {
	// would be nice to have some enpoint to reliably fetch this info
	info, err := GetLatestInfo("en-US")
	if err != nil {
		return GetStaticUAInfo(), err
	}

	return info, nil
}

func GetStaticUAInfo() UserAgentInfo {
	return UserAgentInfo{
		UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36",
		AcceptLanguage: "en-US",
		Platform:       "Win32",
		Metadata: UAMetadata{
			JsHighEntropyHints: JsHighEntropyHints{
				Brands: []Brand{
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
				FullVersionList: []FullVersionList{
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
				UaFullVersion:   "118.0.5993.71",
				Platform:        "Windows",
				PlatformVersion: "6.0.0",
				Architecture:    "x86",
				Model:           "",
				Mobile:          false,
				Bitness:         "64",
				Wow64:           false,
			},
		},
	}
}
