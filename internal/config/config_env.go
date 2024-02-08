package config

import (
	"os"
	"strconv"
	"strings"
)

func (c *Config) env() {
	if api_key := os.Getenv("DATADOG_APIKEY"); api_key != "" {
		c.DataDog.ApiKey = api_key
	}

	if app_key := os.Getenv("DATADOG_APPKEY"); app_key != "" {
		c.DataDog.ApiKey = app_key
	}

	if tag_replicas := os.Getenv("DATADOG_TAG_REPLICAS"); tag_replicas != "" {
		c.DataDog.Tags.Replicas = tag_replicas
	}

	if tag_memory_usage := os.Getenv("DATADOG_TAG_MEMORY_USAGE"); tag_memory_usage != "" {
		c.DataDog.Tags.Memory.Usage = tag_memory_usage
	}

	if tag_memory_request := os.Getenv("DATADOG_TAG_MEMORY_REQUEST"); tag_memory_request != "" {
		c.DataDog.Tags.Memory.Request = tag_memory_request
	}

	if tag_memory_limit := os.Getenv("DATADOG_TAG_MEMORY_LIMIT"); tag_memory_limit != "" {
		c.DataDog.Tags.Memory.Limit = tag_memory_limit
	}

	if tag_cpu_usage := os.Getenv("DATADOG_TAG_CPU_USAGE"); tag_cpu_usage != "" {
		c.DataDog.Tags.Cpu.Usage = tag_cpu_usage
	}

	if tag_cpu_request := os.Getenv("DATADOG_TAG_CPU_REQUEST"); tag_cpu_request != "" {
		c.DataDog.Tags.Cpu.Request = tag_cpu_request
	}

	if tag_cpu_limit := os.Getenv("DATADOG_TAG_CPU_LIMIT"); tag_cpu_limit != "" {
		c.DataDog.Tags.Cpu.Limit = tag_cpu_limit
	}

	if tag_namespace := os.Getenv("DATADOG_TAG_NAMESPACE"); tag_namespace != "" {
		c.DataDog.Tags.Namespace = tag_namespace
	}

	if tag_deployment := os.Getenv("DATADOG_TAG_DEPLOYMENT"); tag_deployment != "" {
		c.DataDog.Tags.Deployment = tag_deployment
	}

	if begin := os.Getenv("DATADOG_TIMES_BEGIN"); begin != "" {
		if val, err := strconv.Atoi(begin); err != nil {
			c.DataDog.Times.Begin = val
		}
	}

	if end := os.Getenv("DATADOG_TIMES_END"); end != "" {
		if val, err := strconv.Atoi(end); err != nil {
			c.DataDog.Times.Begin = val
		}
	}

	if filters := os.Getenv("FILTERS"); filters != "" {
		spl := strings.Split(filters, ",")
		for _, s := range spl {
			spl2 := strings.Split(s, "=")
			if len(spl2) < 2 {
				continue
			}
			c.Filters[spl2[0]] = spl2[1]
		}
	}

	if scales_cpu_request := os.Getenv("SCALES_CPU_REQUEST"); scales_cpu_request != "" {
		if val, err := strconv.ParseFloat(scales_cpu_request, 64); err != nil {
			c.Scales.Cpu.Request = val
		}
	}

	if scales_cpu_limit := os.Getenv("SCALES_CPU_LIMIT"); scales_cpu_limit != "" {
		if val, err := strconv.ParseFloat(scales_cpu_limit, 64); err != nil {
			c.Scales.Cpu.Limit = val
		}
	}

	if scales_memory_request := os.Getenv("SCALES_MEMORY_REQUEST"); scales_memory_request != "" {
		if val, err := strconv.ParseFloat(scales_memory_request, 64); err != nil {
			c.Scales.Memory.Request = val
		}
	}

	if scales_memory_limit := os.Getenv("SCALES_MEMORY_LIMIT"); scales_memory_limit != "" {
		if val, err := strconv.ParseFloat(scales_memory_limit, 64); err != nil {
			c.Scales.Memory.Limit = val
		}
	}
}
