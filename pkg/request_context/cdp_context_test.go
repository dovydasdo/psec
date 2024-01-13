package requestcontext

import (
	"log/slog"
	"testing"
	"time"

	"github.com/dovydasdo/psec/config"
)

func TestNavigate(t *testing.T) {

	cfg := config.ConfCDPLaunch{
		BinPath:       "",
		InjectionPath: "",
	}

	ctx := GetCDPContext(
		cfg,
		slog.Default(),
	)

	ins := NavigateInstruction{
		URL:           "https://google.com",
		DoneCondition: time.Second * 5,
	}

	ctx.Do(
		ins,
	)

}

// TODO: test injection
// TODO: test closing
// TODO: test navigation
// TODO: test proxy
