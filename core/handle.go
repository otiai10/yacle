package core

import (
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	cwl "github.com/otiai10/cwl.go"
)

// Handler ...
type Handler struct {
	Workflow   *cwl.Root
	Parameters cwl.Parameters
	Outdir     string
	Quiet      bool
	Alias      map[string]interface{}
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
		Alias: map[string]interface{}{
			"runtime.cores": "2", // TODO: This is tmp hard coding!!!!
		},
	}, nil
}

// SetLogger ...
func (h *Handler) SetLogger(logger *log.Logger) {
	h.logger = logger
}

// Handle ...
func (h *Handler) Handle(job cwl.Parameters) error {
	h.Parameters = job
	// FIXME: this procedure ONLY adjusts to "baseCommand" job
	arguments := h.ensureArguments()

	priors, inputs, err := h.ensureInputs()
	if err != nil {
		return fmt.Errorf("failed to ensure required inputs: %v", err)
	}

	oneline := append(priors, append(arguments, inputs...)...)

	if len(h.Workflow.BaseCommands) > 1 {
		oneline = append(h.Workflow.BaseCommands[1:], oneline...)
	}
	var cmd *exec.Cmd
	if len(h.Workflow.BaseCommands) != 0 {
		cmd = exec.Command(h.Workflow.BaseCommands[0], oneline...)
	} else {
		// using arguments valueFrom
		cmd = exec.Command("bash", "-c", h.Workflow.Arguments[0].Binding.ValueFrom.Key())
	}
	cmd.Dir = filepath.Dir(h.Workflow.Path)
	if h.Workflow.Outputs[0].Types[0].Type != "File" && h.Workflow.Outputs[0].Types[0].Type != "stdout" && h.Workflow.Outputs[0].Types[0].Type != "stderr" {
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to execute BaseCommand: %v", err)
		}
	}
	filename := h.Workflow.Stdout
	if h.Workflow.Stdout == "" {
		filename = h.Workflow.Stderr
		if h.Workflow.Stderr == "" {
			letters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
			filenamelen := 16
			randombytearray := make([]byte, filenamelen)
			rand.Seed(time.Now().UnixNano())
			for i := range randombytearray {
				randombytearray[i] = letters[rand.Intn(len(letters))]
			}
			filename = string(randombytearray)
		}
	}

	// {{{ TODO: Remove this hard coding!!
	if output, err := os.Open(filepath.Join(filepath.Dir(h.Workflow.Path), "cwl.output.json")); err == nil {
		defer output.Close()
		if _, err := io.Copy(os.Stdout, output); err != nil {
			return fmt.Errorf("failed to dump standard output file: %v", err)
		}
		if err := os.Rename(output.Name(), filepath.Join(h.Outdir, filepath.Base(output.Name()))); err != nil {
			return fmt.Errorf("failed to move starndard output file: %v", err)
		}
	} else if h.Workflow.Outputs[0].Types[0].Type == "File" || h.Workflow.Outputs[0].Types[0].Type == "stdout" || h.Workflow.Outputs[0].Types[0].Type == "stderr" {
		if output, err := os.Create(filepath.Join(filepath.Dir(h.Workflow.Path), filename)); err == nil {
			defer output.Close()
			cmd.Stdout = output
			cmd.Stderr = output

			err = cmd.Start()
			if err != nil {
				panic(err)
			}
			cmd.Wait()
			// if _, err := io.Copy(os.Stdout, output); err != nil {
			// 	return fmt.Errorf("failed to dump standard output file: %v", err)
			// }
			if err := os.Rename(output.Name(), filepath.Join(h.Outdir, filepath.Base(output.Name()))); err != nil {
				return fmt.Errorf("failed to move starndard output file: %v", err)
			}
		}
	}
	if h.Workflow.Outputs[0].Types[0].Type == "File" || h.Workflow.Outputs[0].Types[0].Type == "stdout" || h.Workflow.Outputs[0].Types[0].Type == "stderr" {
		// TODO output file information
		// This is for n7
		// n9 requires extend here.
		// before this part, we need to create given filename
		checksum := ""
		path := ""
		basename := ""
		location := ""
		var size int64 = -1
		if f, err := os.Open(filepath.Join(h.Outdir, filepath.Base(filename))); err == nil {
			path = f.Name()
			basename = filepath.Base(path)
			location = fmt.Sprintf("file://%s", path)

			defer f.Close()
			fileinfo, _ := f.Stat()
			size = fileinfo.Size()
			h := sha1.New()
			if _, err := io.Copy(h, f); err != nil {
				log.Fatal(err)
			}

			checksum = fmt.Sprintf("sha1$%x", string(h.Sum(nil)))
		}
		fmt.Println("{\"output_file\":{\"checksum\": \"" + checksum + "\",\"basename\": \"" + basename + "\",\"location\": \"" + location + "\",\"path\": \"" + path + "\",\"class\": \"File\",\"size\": " + strconv.FormatInt(size, 10) + "}}")
	}
	// }}}

	return nil
}

// ensureArguments ...
func (h *Handler) ensureArguments() []string {
	result := []string{}
	sort.Sort(h.Workflow.Arguments)
	for i, arg := range h.Workflow.Arguments {
		if arg.Binding != nil && arg.Binding.ValueFrom != nil {
			h.Workflow.Arguments[i].Value = h.AliasFor(arg.Binding.ValueFrom.Key())
		}
		result = append(result, h.Workflow.Arguments[i].Flatten()...)
	}
	return result
}

// ensureInputs ...
func (h *Handler) ensureInputs() (priors []string, result []string, err error) {
	defer func() {
		if i := recover(); i != nil {
			err = fmt.Errorf("failed to collate required inputs and provided params: %v", i)
		}
	}()
	sort.Sort(h.Workflow.Inputs)
	for _, in := range h.Workflow.Inputs {
		in, err = h.ensureInput(in)
		if err != nil {
			return priors, result, err
		}
		if in.Binding == nil {
			continue
		}
		if in.Binding.Position < 0 {
			priors = append(priors, in.Flatten()...)
		} else {
			result = append(result, in.Flatten()...)
		}
	}
	return priors, result, nil
}

// ensureInput ...
func (h *Handler) ensureInput(input cwl.Input) (cwl.Input, error) {
	if provided, ok := h.Parameters[input.ID]; ok {
		input.Provided = provided
	}
	if input.Default == nil && input.Binding == nil && input.Provided == nil {
		return input, fmt.Errorf("input `%s` doesn't have default field but not provided", input.ID)
	}
	if key, needed := input.Types[0].NeedRequirement(); needed {
		for _, req := range h.Workflow.Requirements {
			for _, requiredtype := range req.Types {
				if requiredtype.Name == key {
					input.RequiredType = &requiredtype
					input.Requirements = h.Workflow.Requirements
				}
			}
		}
	}
	return input, nil
}

// AliasFor ...
func (h *Handler) AliasFor(key string) string {
	v, ok := h.Alias[key]
	if !ok {
		return ""
	}
	return fmt.Sprintf("%v", v)
}
