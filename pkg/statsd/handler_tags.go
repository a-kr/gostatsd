package statsd

import (
	"context"

	"github.com/atlassian/gostatsd"
)

type TagHandler struct {
	metrics       MetricHandler
	events        EventHandler
	tags          gostatsd.Tags // Tags to add to all metrics
	estimatedTags int
}

// NewTagHandler initialises a new handler which adds unique tags and sends metrics/events to the next handler
func NewTagHandler(metrics MetricHandler, events EventHandler, tags gostatsd.Tags) *TagHandler {
	return &TagHandler{
		metrics:       metrics,
		events:        events,
		tags:          tags,
		estimatedTags: len(tags) + metrics.EstimatedTags(),
	}
}

// EstimatedTags returns a guess for how many tags to pre-allocate
func (th *TagHandler) EstimatedTags() int {
	return th.estimatedTags
}

// DispatchMetric adds the unique tags from the TagHandler to the metric and passes it to the next stage in the pipeline
func (th *TagHandler) DispatchMetric(ctx context.Context, m *gostatsd.Metric) error {
	if m.Hostname == "" {
		m.Hostname = string(m.SourceIP)
	}
	m.Tags = uniqueTags(m.Tags, th.tags)
	return th.metrics.DispatchMetric(ctx, m)
}

// DispatchEvent adds the unique tags from the TagHandler to the event and passes it to the next stage in the pipeline
func (th *TagHandler) DispatchEvent(ctx context.Context, e *gostatsd.Event) error {
	if e.Hostname == "" {
		e.Hostname = string(e.SourceIP)
	}
	e.Tags = uniqueTags(e.Tags, th.tags)
	return th.events.DispatchEvent(ctx, e)
}

// WaitForEvents waits for all event-dispatching goroutines to finish.
func (th *TagHandler) WaitForEvents() {
	th.events.WaitForEvents()
}

func uniqueTags(t1 gostatsd.Tags, t2 gostatsd.Tags) gostatsd.Tags {
	tags := gostatsd.Tags{}
	seen := map[string]bool{}

	for _, v := range t1 {
		if !seen[v] {
			tags = append(tags, v)
			seen[v] = true
		}
	}

	for _, v := range t2 {
		if !seen[v] {
			tags = append(tags, v)
			seen[v] = true
		}
	}

	return tags
}
