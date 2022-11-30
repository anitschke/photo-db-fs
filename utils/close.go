package utils

import (
	"fmt"
	"io"

	"go.uber.org/zap"
)

func CloseAndLogErrors(c io.Closer) {
	if err := c.Close(); err != nil {
		zap.L().Error("error closing", zap.Any("closer", c), zap.Error(err))
	}
}

func CloseAndPanicOnError(c io.Closer) {
	if err := c.Close(); err != nil {
		panic(fmt.Errorf("error closing: %w", err))
	}
}
