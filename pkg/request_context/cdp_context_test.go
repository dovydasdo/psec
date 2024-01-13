package requestcontext

import (
	"log/slog"
	"os"
	"testing"

	"github.com/dovydasdo/psec/config"
)

type leveler struct {
}

func (leveler) Level() slog.Level {
	return -4
}

func TestNavigate(t *testing.T) {
	// cfg := config.ConfCDPLaunch{
	// 	BinPath:       "/home/ddom/Documents/cshell/chrome/chrome-linux64/chrome-headless-shell",
	// 	InjectionPath: "/home/ddom/coding/psec/util/injections/stealth.min.js",
	// }
	cfg := config.ConfCDPLaunch{
		BinPath:       "/usr/bin/google-chrome",
		InjectionPath: "/home/ddom/coding/psec/util/injections/stealth.min.js",
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: leveler{}}))

	ctx := GetCDPContext(
		cfg,
		logger,
	)

	ctx.Initialize()
	defer ctx.Close()

	testUrl := "https://news.ycombinator.com/"

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
			if val.Response.URL == testUrl {
				isInState = true
			}
		}

		return true
	})

	t.Logf("is in state: %v", isInState)

	if !isInState {
		t.Errorf("url %v was not found in state after performing the request", testUrl)
	}
}

// TODO: test injection
// TODO: test closing
// TODO: test navigation
// TODO: test proxy
