package fixes

import "log/slog"

func sensitive() {
	token := "x"
	slog.Info("token " + token) // want `message contains sensitive keyword "token" from identifier "token"`
}
