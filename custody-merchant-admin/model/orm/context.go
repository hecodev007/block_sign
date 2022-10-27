package orm

import (
	"github.com/opentracing/opentracing-go"
)

type context interface {
	OpenTracingSpan() opentracing.Span
}
