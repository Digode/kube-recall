package main

import (
	"k8s-resources-update/internal/checker"
	"k8s-resources-update/internal/config"
	"k8s-resources-update/internal/util"
)

var logger = util.GetLogger()
var cfg = config.Get()

func main() {
	logger.Info("Starting the application...")

	checker.CheckResources()

	logger.Info("Checking resources...")
}
