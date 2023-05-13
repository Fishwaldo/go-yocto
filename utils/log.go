package utils

import (
	"github.com/pterm/pterm"
)

var Logger *pterm.Logger

func InitLogger() {
	Logger = pterm.DefaultLogger.WithLevel(pterm.LogLevelTrace) // Only show logs with a level of Trace or higher.
}