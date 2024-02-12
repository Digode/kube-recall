package main

import (
	"kube-recall/internal/checker"
	"kube-recall/internal/config"
	"kube-recall/internal/util"
)

var logger = util.GetLogger()
var cfg = config.Get()

func main() {
	logger.Debug("Starting the application...")

	checker.CheckResources()

	logger.Debug("Checking resources...")
}
