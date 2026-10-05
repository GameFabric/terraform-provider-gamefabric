package mps

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

const (
	// MetricPortPolicy is the TFP-only port policy for Prometheus metrics scraping.
	// The TFP translates it to Agones None before sending to the GameFabric API and
	// injects the required gameserver labels and annotations for scraping.
	MetricPortPolicy = "Metric"

	// MetricsScrapeLabel is the gameserver label that enables Prometheus scraping.
	MetricsScrapeLabel = "g8c.io/gameserver-scrape"

	// MetricsEndpointsAnnotation is the gameserver annotation that lists all metrics
	// endpoints in the format "port=path,port=path".
	MetricsEndpointsAnnotation = "g8c.io/metrics-endpoints"

	// defaultMetricsPath is used when a Metric port does not specify a path.
	defaultMetricsPath = "/metrics"
)

// HasMetricPorts returns true if any container in the slice contains a port with
// policy Metric.
func HasMetricPorts(containers []ContainerModel) bool {
	for _, c := range containers {
		for _, p := range c.Ports {
			if p.Policy.ValueString() == MetricPortPolicy {
				return true
			}
		}
	}
	return false
}

// MetricsAnnotationValue builds the comma-separated "port=path" string for all
// Metric ports across all containers, sorted by port for deterministic output.
func MetricsAnnotationValue(containers []ContainerModel) string {
	type entry struct {
		port uint16
		path string
	}

	var entries []entry
	for _, c := range containers {
		for _, p := range c.Ports {
			if p.Policy.ValueString() != MetricPortPolicy {
				continue
			}
			path := defaultMetricsPath
			if p.Path.ValueString() != "" {
				path = p.Path.ValueString()
			}
			entries = append(entries, entry{uint16(p.ContainerPort.ValueInt32()), path})
		}
	}

	sort.Slice(entries, func(i, j int) bool { return entries[i].port < entries[j].port })

	parts := make([]string, len(entries))
	for i, e := range entries {
		parts[i] = fmt.Sprintf("%d=%s", e.port, e.path)
	}
	return strings.Join(parts, ",")
}

// ParseMetricsAnnotation parses "port=path,port=path" into a map of port → path.
// Returns nil if the annotation is empty.
func ParseMetricsAnnotation(annotation string) map[uint16]string {
	if annotation == "" {
		return nil
	}
	result := make(map[uint16]string)
	for pair := range strings.SplitSeq(annotation, ",") {
		pair = strings.TrimSpace(pair)
		portStr, path, ok := strings.Cut(pair, "=")
		if !ok {
			continue
		}
		port, err := strconv.ParseUint(portStr, 10, 16)
		if err != nil {
			continue
		}
		result[uint16(port)] = path
	}
	return result
}
