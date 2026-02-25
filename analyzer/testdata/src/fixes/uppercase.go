package fixes

import "log/slog"

func uppercase() {
	slog.Info("Bad start") // want `message "Bad start" starts with the capital letter`
}
