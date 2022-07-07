package handler

import (
	"github.com/ispras/michman/internal/protobuf"
)

type ServiceExists struct {
	Exists  bool
	Service *protobuf.Service
}

const (
	QueryViewTypeFull    = "full"
	QueryViewTypeSummary = "summary"
	QueryViewKey         = "view"
)
