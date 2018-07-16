package core

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/robertkrimen/otto"

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
	// Set command execution directory
	commandExecDir := h.Workflow.Path
	cmd.Dir = filepath.Dir(commandExecDir)
	if h.Workflow.Outputs[0].Types[0].Type == "Directory" {
		// Create TempDir for exec
		commandExecDir, _ = ioutil.TempDir("/tmp", "yacleexec")
		srcName := filepath.Join(filepath.Dir(h.Workflow.Path), inputs[0])
		dstName := filepath.Join(commandExecDir, inputs[0])

		// Create copied file
		os.Link(srcName, dstName)
		cmd.Dir = commandExecDir
	}
	if h.Workflow.Outputs[0].Types[0].Type != "File" && h.Workflow.Outputs[0].Types[0].Type != "int" && h.Workflow.Outputs[0].Types[0].Type != "stdout" && h.Workflow.Outputs[0].Types[0].Type != "stderr" && h.Workflow.Outputs[0].Types[0].Type != "Directory" {
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to execute BaseCommand: %v", err)
		}
	}
	var filename string
	isStdoutFlag := false
	isStderrFlag := false
	stderrFilename := h.Workflow.Stderr
	for i := range h.Workflow.Outputs {
		switch h.Workflow.Outputs[i].Types[0].Type {
		case "stderr":
			stderrFilename = h.Workflow.Outputs[i].ID
			isStderrFlag = true
		case "stdout":
			filename = h.Workflow.Outputs[i].ID
			isStdoutFlag = true
		}
	}
	if h.Workflow.Stdout != "" {
		filename = h.Workflow.Stdout
		isStdoutFlag = true
	}
	if h.Workflow.Stderr != "" {
		stderrFilename = h.Workflow.Stderr
		isStderrFlag = true
	}
	// TODO check filename
	if filename == "" {
		filename = stderrFilename
		if stderrFilename == "" {
			letters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
			filenamelen := 16
			randombytearray := make([]byte, filenamelen)
			rand.Seed(time.Now().UnixNano())
			for i := range randombytearray {
				randombytearray[i] = letters[rand.Intn(len(letters))]
			}
			filename = string(randombytearray)
			// TODO only accept
			isStdoutFlag = true
		}
	}

	// {{{ TODO: Remove this hard coding!!
	if output, err := os.Open(filepath.Join(filepath.Dir(commandExecDir), "cwl.output.json")); err == nil {
		defer output.Close()
		if _, err := io.Copy(os.Stdout, output); err != nil {
			return fmt.Errorf("failed to dump standard output file: %v", err)
		}
		if err := os.Rename(output.Name(), filepath.Join(h.Outdir, filepath.Base(output.Name()))); err != nil {
			return fmt.Errorf("failed to move starndard output file: %v", err)
		}
	} else if h.Workflow.Outputs[0].Types[0].Type == "File" || h.Workflow.Outputs[0].Types[0].Type == "int" || h.Workflow.Outputs[0].Types[0].Type == "stdout" || h.Workflow.Outputs[0].Types[0].Type == "stderr" || h.Workflow.Outputs[0].Types[0].Type == "Directory" {
		var output *os.File
		var err error
		if isStdoutFlag {
			output, err = os.Create(filepath.Join(filepath.Dir(commandExecDir), filename))
		}
		var stderrOutput *os.File
		var stderrErr error
		if isStderrFlag {
			stderrOutput, stderrErr = os.Create(filepath.Join(filepath.Dir(commandExecDir), stderrFilename))
		} else {
			stderrOutput = output
		}
		if err == nil {
			defer output.Close()
		}
		if stderrErr == nil {
			defer stderrOutput.Close()
		}
		if err == nil {
			cmd.Stdout = output
			cmd.Stderr = stderrOutput
			vm := otto.New()
			inputs, _ := vm.Object(`inputs = {}`)
			file1, _ := vm.Object(`file1 = {}`)
			inputs.Set("file1", file1)
			file1.Set("path", "whale.txt")
			stdinFile, _ := vm.Run("inputs.file1.path")
			sF, _ := os.Open(filepath.Join(filepath.Dir(commandExecDir) + "/" + stdinFile.String()))
			stdin, errStdin := cmd.StdinPipe()
			if errStdin != nil {
				fmt.Println(errStdin)
			}

			io.Copy(stdin, bufio.NewReader(sF))
			stdin.Close()
			err = cmd.Start()
			if err != nil {
				panic(err)
			}
			cmd.Wait()
			// if _, err := io.Copy(os.Stdout, output); err != nil {
			// 	return fmt.Errorf("failed to dump standard output file: %v", err)
			// }
			if isStdoutFlag {
				if err := os.Rename(output.Name(), filepath.Join(h.Outdir, filepath.Base(output.Name()))); err != nil {
					return fmt.Errorf("failed to move starndard output file: %v", err)
				}
			}
			if isStderrFlag {
				if stderrErr := os.Rename(stderrOutput.Name(), filepath.Join(h.Outdir, filepath.Base(stderrOutput.Name()))); stderrErr != nil {
					return fmt.Errorf("failed to move starndard error file: %v", err)
				}
			}
		}
	}
	if h.Workflow.Outputs[0].Types[0].Type == "File" || h.Workflow.Outputs[0].Types[0].Type == "int" || h.Workflow.Outputs[0].Types[0].Type == "stdout" || h.Workflow.Outputs[0].Types[0].Type == "stderr" || h.Workflow.Outputs[0].Types[0].Type == "Directory" {
		// TODO output file information
		// This is for n7
		// n9 requires extend here.
		// before this part, we need to create given filename
		if h.Workflow.Outputs[0].Types[0].Type != "Directory" &&
			!(isStdoutFlag == true && isStderrFlag == true) {
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
			outputIdentifier := h.Workflow.Outputs[0].ID
			if h.Workflow.Outputs[0].Binding != nil && h.Workflow.Outputs[0].Binding.LoadContents {
				if f2, err := os.Open(filepath.Join(h.Outdir, filepath.Base(filename))); err == nil {
					defer f2.Close()
					lr := io.LimitReader(f2, 64*1024)
					buf := new(bytes.Buffer)
					buf.ReadFrom(lr)
					s := buf.String()
					vm2 := otto.New()
					self := []map[string]interface{}{
						{
							"contents": s,
						},
					}
					vm2.Set("self", self)
					resultObject, _ := vm2.Run("parseInt(self[0].contents)")
					result, _ := resultObject.ToString()
					fmt.Println("{\"" + outputIdentifier + "\": " + result + "}")
				}
			} else {
				fmt.Println("{\"" + outputIdentifier + "\":{\"checksum\": \"" + checksum + "\",\"basename\": \"" + basename + "\",\"location\": \"" + location + "\",\"path\": \"" + path + "\",\"class\": \"File\",\"size\": " + strconv.FormatInt(size, 10) + "}}")
			}
		} else {
			if isStdoutFlag == true && isStderrFlag == true {
				files, err := ioutil.ReadDir(h.Outdir)
				if err != nil {
					panic(err)
				}
				listingString := ""
				for _, file := range files {
					checksum := ""
					path := ""
					basename := ""
					var size int64 = -1
					if f, err := os.Open(h.Outdir + "/" + file.Name()); err == nil {
						path = f.Name()
						basename = filepath.Base(path)

						defer f.Close()
						fileinfo, _ := f.Stat()
						size = fileinfo.Size()
						h := sha1.New()
						if _, err := io.Copy(h, f); err != nil {
							log.Fatal(err)
						}

						checksum = fmt.Sprintf("sha1$%x", string(h.Sum(nil)))
						if listingString != "" {
							listingString = listingString + ","
						}
						listingString = listingString + "\"" + basename + "\":{\"checksum\": \"" + checksum + "\",\"location\": \"" + "Any" + "\",\"class\": \"File\",\"size\": " + strconv.FormatInt(size, 10) + "}"
					}
				}
				fmt.Println("{" + listingString + "}")
			} else if h.Workflow.Outputs[0].Binding.Glob[0] == "." {
				// Remove linked input file
				dstName := filepath.Join(commandExecDir, inputs[0])
				errRemove := os.Remove(dstName)

				if errRemove != nil {
					panic(errRemove)
				}
				files, err := ioutil.ReadDir(commandExecDir)
				if err != nil {
					panic(err)
				}
				listingString := ""
				for _, file := range files {
					checksum := ""
					path := ""
					basename := ""
					location := ""
					var size int64 = -1
					if f, err := os.Open(commandExecDir + "/" + file.Name()); err == nil {
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
						if listingString != "" {
							listingString = listingString + ","
						}
						listingString = listingString + "{\"checksum\": \"" + checksum + "\",\"basename\": \"" + basename + "\",\"location\": \"" + location + "\",\"path\": \"" + path + "\",\"class\": \"File\",\"size\": " + strconv.FormatInt(size, 10) + "}"
					}
				}
				outputIdentifier := h.Workflow.Outputs[0].ID
				fmt.Println("{\"" + outputIdentifier + "\":{\"class\": \"Directory\",\"listing\": [" + listingString + "]}}")
			}
		}
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
