package core

import "strings"

// Logf ...
func (h *Handler) Logf(format string, v ...interface{}) {
	if h.logger == nil {
		return
	}
	if h.Quiet {
		return
	}
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	h.logger.Printf(format, v...)
}

// Logln ...
func (h *Handler) Logln(v ...interface{}) {
	if h.logger == nil {
		return
	}
	h.logger.Println(v...)
}
