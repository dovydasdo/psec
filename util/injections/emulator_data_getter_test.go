package util

import (
	"strconv"
	"strings"
	"testing"
)

func TestGetClientHints(t *testing.T) {
	version := 120

	info, err := GetClientHints(version)
	if err != nil {
		t.Errorf("failed at getting client hints: %v", err)
	}

	gotVersion, err := strconv.ParseInt(strings.Split(info.JsHighEntropyHints.UaFullVersion, ".")[0], 10, 64)
	if err != nil {
		t.Errorf("failed to parse received ua version: %v", err)
	}

	if int(gotVersion) != version {
		t.Logf("version from provider does not match the desired verison, vanted: %v, got %v", version, version)
		t.Fail()
	}
}
