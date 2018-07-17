package core

import (
	"github.com/otiai10/yacle/core/class"
)

// Tool ...
type Tool interface {
	Run() error
	Finalize() error
}

// ClassTool constructs and initializes ClassTool, e.g. CommandLineTool.
func (h *Handler) ClassTool() (Tool, error) {
	switch h.Workflow.Class {
	case "CommandLineTool":
		return &class.CommandLineTool{
			Outdir:     h.Outdir,
			Root:       h.Workflow,
			Parameters: h.Parameters,
		}, nil
	default:
		return &class.CommandLineTool{
			Outdir:     h.Outdir,
			Root:       h.Workflow,
			Parameters: h.Parameters,
		}, nil
	}
}
