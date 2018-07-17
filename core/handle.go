package core

import (
	"log"
	"os"

	cwl "github.com/otiai10/cwl.go"
)

// Handler ...
type Handler struct {
	Workflow   *cwl.Root
	Parameters cwl.Parameters
	Outdir     string
	Quiet      bool
	logger     *log.Logger
}

// NewHandler ...
func NewHandler(root *cwl.Root) (*Handler, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return &Handler{
		Workflow: root,
		Outdir:   cwd,
	}, nil
}

// SetLogger ...
func (h *Handler) SetLogger(logger *log.Logger) {
	h.logger = logger
}

// Handle is an entrypoint of the engine for CWL.
func (h *Handler) Handle(job cwl.Parameters) error {

	h.Parameters = job

	tool, err := h.ClassTool()
	if err != nil {
		return err
	}

	if err := tool.Run(); err != nil {
		return err
	}

	return nil
}
