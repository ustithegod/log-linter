package fixes

import "log/slog"

func nonEnglish() {
	slog.Info("hello мир") // want `message "hello мир" contains non-english letters`
}
