package zapdriver

import (
	"fmt"
	"go.uber.org/zap"
	"regexp"
	"strconv"
)

const (
	traceKey = "logging.googleapis.com/trace"
	spanKey = "logging.googleapis.com/spanId"
	traceSampledKey = "logging.googleapis.com/trace_sampled"
)

// TraceContext adds the correct Stackdriver "trace", "span", "trace_sampled fields
//
// see: https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry
func TraceContext(traceContext string, projectName string) []zap.Field {
	r := regexp.MustCompile(`(?P<Trace>[0-9a-zA-Z]+)/(?P<Span>[0-9a-zA-Z]+);o=(?P<TraceSampled>[0-1])`)
	matches := r.FindStringSubmatch(traceContext)
	if len(matches) == 4 {
		return []zap.Field{
			zap.String(traceKey, fmt.Sprintf("projects/%s/traces/%s", projectName, matches[1])),
			zap.String(spanKey, matches[2]),
			zap.String(traceSampledKey, strconv.FormatBool(matches[3] == "1")),
		}
	}
	return []zap.Field{}
}
