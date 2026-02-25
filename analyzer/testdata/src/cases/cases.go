package cases

import (
	"context"
	"log/slog"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func slogCases() {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(nil, nil))

	logger.Info("Bad start") // want `message "Bad start" starts with the capital letter`
	logger.Info("hello")
	logger.Error("ошибка") // want `message "ошибка" contains non-english letters`
	logger.Warn("boom!!!") // want `message "boom!!!" contains special symbols`
	logger.InfoContext(ctx, "Bad ctx") // want `message "Bad ctx" starts with the capital letter`
	logger.Log(ctx, slog.LevelInfo, "Bad log") // want `message "Bad log" starts with the capital letter`

	slog.Info("Bad pkg") // want `message "Bad pkg" starts with the capital letter`
	slog.Warn("warn 🚀") // want `message "warn 🚀" contains special symbols`
}

func zapCases() {
	logger, _ := zap.NewProduction()

	logger.Info("Bad zap") // want `message "Bad zap" starts with the capital letter`
	logger.Info("ok")
	logger.Warn("warn 🚀") // want `message "warn 🚀" contains special symbols`
	logger.Error("ошибка") // want `message "ошибка" contains non-english letters`
	logger.Log(zapcore.InfoLevel, "Bad log") // want `message "Bad log" starts with the capital letter`
}

func sensitiveCases() {
	logger := slog.Default()

	password := "x"
	user_token := "y"
	authToken := "z"
	sessionID := "sid"

	logger.Info("password " + password) // want `message contains sensitive keyword "password" from identifier "password"`
	logger.Info("token " + user_token) // want `message contains sensitive keyword "token" from identifier "user_token"`
	logger.Info("auth " + authToken) // want `message contains sensitive keyword "token" from identifier "authToken"`
	logger.Info("session " + sessionID)
}
