package datadog

import (
	"context"
	"kube-recall/internal/config"
	"kube-recall/internal/util"
	"strings"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
)

var cfg = config.Get()

func getContext() context.Context {
	return context.WithValue(context.Background(), datadog.ContextAPIKeys, map[string]datadog.APIKey{
		"apiKeyAuth": {
			Key: cfg.Provider.DataDog.ApiKey,
		},
		"appKeyAuth": {
			Key: cfg.Provider.DataDog.AppKey,
		},
	})
}

func getTag(tags []string, prefix string) string {
	prefix += ":"
	for _, tag := range tags {
		if strings.HasPrefix(tag, prefix) {
			return strings.TrimPrefix(tag, prefix)
		}
	}
	return ""
}

func timeSpampToTime(timestamp int64) time.Time {
	sec := timestamp / 1000
	nsec := (timestamp % 1000) * 1e6
	return time.Unix(sec, nsec)
}

func bytesToMb(b float64) float64 {
	return util.Round(b/1024/1024, 2)
}

func nanoToCore(b float64) float64 {
	return util.Round(b/1000/1000, 2)
}

func clockToMilicore(b float64) float64 {
	return util.Round(b*1000, 2)
}
