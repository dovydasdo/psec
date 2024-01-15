package requestcontext

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/dovydasdo/psec/config"
)

type leveler struct {
}

func (leveler) Level() slog.Level {
	return -4
}

const ServerPort = 3000

func TestNavigate(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "<html><body><h1>Test</h1></body></html>")
	}))
	defer ts.Close()

	cfg := config.NewCDPLaunchConf()
	if cfg == nil {
		t.Fatalf("failed to read config from env variables")
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: leveler{}}))

	ctx := GetCDPContext(
		*cfg,
		logger,
	)

	ctx.Initialize()
	defer ctx.Close()

	testUrl := ts.URL

	ins := NavigateInstruction{
		URL:           testUrl,
		DoneCondition: DoneResponseReceived(testUrl),
		Filters:       make([]string, 0),
	}

	result, err := ctx.Do(
		ins,
	)

	if err != err {
		t.Errorf("failed to navigate: %v", err)
	}

	for _, res := range result {
		if res.Error != nil {
			t.Errorf("failed to perform instruction: %v, err: %v", res.Type, err)
		}
	}

	state := ctx.GetState()
	isInState := false

	state.NetworkEvents.Range(func(key, value interface{}) bool {
		if val, ok := value.(*NetworkEvent); ok {
			if val.Response.URL == testUrl || val.Response.URL == testUrl+"/" {
				isInState = true
			}
		}

		return true
	})

	if !isInState {
		t.Errorf("url %v was not found in state after performing the request", testUrl)
	}
}

func TestEmulation(t *testing.T) {
	cfg := config.NewCDPLaunchConf()
	if cfg == nil {
		t.Fatalf("failed to read config from env variables")
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: leveler{}}))

	ctx := GetCDPContext(
		*cfg,
		logger,
	)

	ctx.Initialize()
	defer ctx.Close()

	var evalRes bool

	ins := JSEvalInstruction{
		Script:  "navigator.webdriver",
		Timeout: time.Second,
		Result:  &evalRes,
	}

	result, err := ctx.Do(ins)
	if err != nil {
		t.Errorf("failed to eval script: %v", err)
	}

	for _, res := range result {
		if res.Type == "js_eval" {
			if v, ok := res.Value.(*bool); ok {
				if *v != false {
					t.Log("navigator is on")
					t.Fail()
				}

			}
		}
	}

}

// TODO: test proxy
