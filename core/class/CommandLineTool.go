package class

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	cwl "github.com/otiai10/cwl.go"
)

// CommandLineTool represents class described as "CommandLineTool".
type CommandLineTool struct {
	Outdir     string // Given by context
	Root       *cwl.Root
	Parameters cwl.Parameters
	Command    *exec.Cmd
}

// Run ...
func (tool *CommandLineTool) Run() error {

	// FIXME: this procedure ONLY adjusts to "baseCommand" job
	arguments := tool.ensureArguments()

	priors, inputs, err := tool.ensureInputs()
	if err != nil {
		return fmt.Errorf("failed to ensure required inputs: %v", err)
	}

	cmd, err := tool.generateCommand(priors, arguments, inputs)
	if err != nil {
		return fmt.Errorf("failed to generate command struct: %v", err)
	}
	tool.Command = cmd

	if err := tool.cmdが実行されるディレクトリを決める(); err != nil {
		return fmt.Errorf("cmdが実行されるディレクトリを決める: %v", err)
	}

	if err := tool.標準出力の行き先を決める(); err != nil {
		return fmt.Errorf("標準出力の行き先を決める: %v", err)
	}

	if err := tool.標準エラーの行き先を決める(); err != nil {
		return fmt.Errorf("標準エラーの行き先を決める: %v", err)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("exec failed: %v", err)
	}

	if err := tool.outdirを決定する(); err != nil {
		return fmt.Errorf("outdirを決定する: %v", err)
	}

	if err := tool.outdirにいろいろ揃える(); err != nil {
		return fmt.Errorf("outdirにいろいろ揃える: %v", err)
	}

	return nil

	// // Set command execution directory
	// commandExecDir := tool.Root.Path
	// cmd.Dir = filepath.Dir(commandExecDir)
	// if tool.Root.Outputs[0].Types[0].Type == "Directory" {
	// 	// Create TempDir for exec
	// 	commandExecDir, _ = ioutil.TempDir("/tmp", "yacleexec")
	// 	srcName := filepath.Join(filepath.Dir(tool.Root.Path), inputs[0])
	// 	dstName := filepath.Join(commandExecDir, inputs[0])

	// 	// Create copied file
	// 	os.Link(srcName, dstName)
	// 	cmd.Dir = commandExecDir
	// }
	// if tool.Root.Outputs[0].Types[0].Type != "File" && tool.Root.Outputs[0].Types[0].Type != "int" && tool.Root.Outputs[0].Types[0].Type != "stdout" && tool.Root.Outputs[0].Types[0].Type != "stderr" && tool.Root.Outputs[0].Types[0].Type != "Directory" {
	// 	if err := cmd.Run(); err != nil {
	// 		return fmt.Errorf("failed to execute BaseCommand: %v", err)
	// 	}
	// }
	// var filename string
	// isStdoutFlag := false
	// isStderrFlag := false
	// stderrFilename := tool.Root.Stderr
	// for i := range tool.Root.Outputs {
	// 	switch tool.Root.Outputs[i].Types[0].Type {
	// 	case "stderr":
	// 		stderrFilename = tool.Root.Outputs[i].ID
	// 		isStderrFlag = true
	// 	case "stdout":
	// 		filename = tool.Root.Outputs[i].ID
	// 		isStdoutFlag = true
	// 	}
	// }
	// if tool.Root.Stdout != "" {
	// 	filename = tool.Root.Stdout
	// 	isStdoutFlag = true
	// }
	// if tool.Root.Stderr != "" {
	// 	stderrFilename = tool.Root.Stderr
	// 	isStderrFlag = true
	// }
	// // TODO check filename
	// if filename == "" {
	// 	filename = stderrFilename
	// 	if stderrFilename == "" {
	// 		letters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	// 		filenamelen := 16
	// 		randombytearray := make([]byte, filenamelen)
	// 		rand.Seed(time.Now().UnixNano())
	// 		for i := range randombytearray {
	// 			randombytearray[i] = letters[rand.Intn(len(letters))]
	// 		}
	// 		filename = string(randombytearray)
	// 		// TODO only accept
	// 		isStdoutFlag = true
	// 	}
	// }

	// // {{{ TODO: Remove this hard coding!!
	// if output, err := os.Open(filepath.Join(filepath.Dir(commandExecDir), "cwl.output.json")); err == nil {
	// 	defer output.Close()
	// 	if _, err := io.Copy(os.Stdout, output); err != nil {
	// 		return fmt.Errorf("failed to dump standard output file: %v", err)
	// 	}
	// 	if err := os.Rename(output.Name(), filepath.Join(tool.Outdir, filepath.Base(output.Name()))); err != nil {
	// 		return fmt.Errorf("failed to move starndard output file: %v", err)
	// 	}
	// 	return nil
	// }

	// if tool.Root.Outputs[0].Types[0].Type == "File" || tool.Root.Outputs[0].Types[0].Type == "int" || tool.Root.Outputs[0].Types[0].Type == "stdout" || tool.Root.Outputs[0].Types[0].Type == "stderr" || tool.Root.Outputs[0].Types[0].Type == "Directory" {
	// 	var output *os.File
	// 	var err error
	// 	if isStdoutFlag {
	// 		output, err = os.Create(filepath.Join(filepath.Dir(commandExecDir), filename))
	// 	}
	// 	var stderrOutput *os.File
	// 	var stderrErr error
	// 	if isStderrFlag {
	// 		stderrOutput, stderrErr = os.Create(filepath.Join(filepath.Dir(commandExecDir), stderrFilename))
	// 	} else {
	// 		stderrOutput = output
	// 	}
	// 	if err == nil {
	// 		defer output.Close()
	// 	}
	// 	if stderrErr == nil {
	// 		defer stderrOutput.Close()
	// 	}
	// 	if err == nil {
	// 		cmd.Stdout = output
	// 		cmd.Stderr = stderrOutput
	// 		vm := otto.New()
	// 		inputs, _ := vm.Object(`inputs = {}`)
	// 		file1, _ := vm.Object(`file1 = {}`)
	// 		inputs.Set("file1", file1)
	// 		file1.Set("path", "whale.txt")
	// 		stdinFile, _ := vm.Run("inputs.file1.path")
	// 		sF, _ := os.Open(filepath.Join(filepath.Dir(commandExecDir) + "/" + stdinFile.String()))
	// 		stdin, errStdin := cmd.StdinPipe()
	// 		if errStdin != nil {
	// 			fmt.Println(errStdin)
	// 		}

	// 		io.Copy(stdin, bufio.NewReader(sF))
	// 		stdin.Close()
	// 		if err := cmd.Run(); err != nil {
	// 			return err
	// 		}
	// 		// if _, err := io.Copy(os.Stdout, output); err != nil {
	// 		// 	return fmt.Errorf("failed to dump standard output file: %v", err)
	// 		// }
	// 		if isStdoutFlag {
	// 			if err := os.Rename(output.Name(), filepath.Join(tool.Outdir, filepath.Base(output.Name()))); err != nil {
	// 				return fmt.Errorf("failed to move starndard output file: %v", err)
	// 			}
	// 		}
	// 		if isStderrFlag {
	// 			if stderrErr := os.Rename(stderrOutput.Name(), filepath.Join(tool.Outdir, filepath.Base(stderrOutput.Name()))); stderrErr != nil {
	// 				return fmt.Errorf("failed to move starndard error file: %v", err)
	// 			}
	// 		}
	// 	}

	// 	// TODO output file information
	// 	// This is for n7
	// 	// n9 requires extend here.
	// 	// before this part, we need to create given filename
	// 	if tool.Root.Outputs[0].Types[0].Type != "Directory" &&
	// 		!(isStdoutFlag == true && isStderrFlag == true) {
	// 		checksum := ""
	// 		path := ""
	// 		basename := ""
	// 		location := ""
	// 		var size int64 = -1
	// 		if f, err := os.Open(filepath.Join(tool.Outdir, filepath.Base(filename))); err == nil {
	// 			path = f.Name()
	// 			basename = filepath.Base(path)
	// 			location = fmt.Sprintf("file://%s", path)

	// 			defer f.Close()
	// 			fileinfo, _ := f.Stat()
	// 			size = fileinfo.Size()
	// 			h := sha1.New()
	// 			if _, err := io.Copy(h, f); err != nil {
	// 				log.Fatal(err)
	// 			}

	// 			checksum = fmt.Sprintf("sha1$%x", string(h.Sum(nil)))
	// 		}
	// 		outputIdentifier := tool.Root.Outputs[0].ID
	// 		if tool.Root.Outputs[0].Binding != nil && tool.Root.Outputs[0].Binding.LoadContents {
	// 			if f2, err := os.Open(filepath.Join(tool.Outdir, filepath.Base(filename))); err == nil {
	// 				defer f2.Close()
	// 				lr := io.LimitReader(f2, 64*1024)
	// 				buf := new(bytes.Buffer)
	// 				buf.ReadFrom(lr)
	// 				s := buf.String()
	// 				vm2 := otto.New()
	// 				self := []map[string]interface{}{
	// 					{
	// 						"contents": s,
	// 					},
	// 				}
	// 				vm2.Set("self", self)
	// 				//
	// 				contentsBytes := []byte(tool.Root.Outputs[0].Binding.Eval)
	// 				contents := string(contentsBytes[2 : len(contentsBytes)-1])
	// 				resultObject, _ := vm2.Run(contents)
	// 				result, _ := resultObject.ToString()
	// 				fmt.Println("{\"" + outputIdentifier + "\": " + result + "}")
	// 			}
	// 		} else {
	// 			fmt.Println("{\"" + outputIdentifier + "\":{\"checksum\": \"" + checksum + "\",\"basename\": \"" + basename + "\",\"location\": \"" + location + "\",\"path\": \"" + path + "\",\"class\": \"File\",\"size\": " + strconv.FormatInt(size, 10) + "}}")
	// 		}
	// 	} else {
	// 		if isStdoutFlag == true && isStderrFlag == true {
	// 			files, err := ioutil.ReadDir(tool.Outdir)
	// 			if err != nil {
	// 				panic(err)
	// 			}
	// 			listingString := ""
	// 			for _, file := range files {
	// 				checksum := ""
	// 				path := ""
	// 				basename := ""
	// 				var size int64 = -1
	// 				if f, err := os.Open(tool.Outdir + "/" + file.Name()); err == nil {
	// 					path = f.Name()
	// 					basename = filepath.Base(path)

	// 					defer f.Close()
	// 					fileinfo, _ := f.Stat()
	// 					size = fileinfo.Size()
	// 					h := sha1.New()
	// 					if _, err := io.Copy(h, f); err != nil {
	// 						log.Fatal(err)
	// 					}

	// 					checksum = fmt.Sprintf("sha1$%x", string(h.Sum(nil)))
	// 					if listingString != "" {
	// 						listingString = listingString + ","
	// 					}
	// 					listingString = listingString + "\"" + basename + "\":{\"checksum\": \"" + checksum + "\",\"location\": \"" + "Any" + "\",\"class\": \"File\",\"size\": " + strconv.FormatInt(size, 10) + "}"
	// 				}
	// 			}
	// 			fmt.Println("{" + listingString + "}")
	// 		} else if tool.Root.Outputs[0].Binding.Glob[0] == "." {
	// 			// Remove linked input file
	// 			dstName := filepath.Join(commandExecDir, inputs[0])
	// 			errRemove := os.Remove(dstName)

	// 			if errRemove != nil {
	// 				panic(errRemove)
	// 			}
	// 			files, err := ioutil.ReadDir(commandExecDir)
	// 			if err != nil {
	// 				panic(err)
	// 			}
	// 			listingString := ""
	// 			for _, file := range files {
	// 				checksum := ""
	// 				path := ""
	// 				basename := ""
	// 				location := ""
	// 				var size int64 = -1
	// 				if f, err := os.Open(commandExecDir + "/" + file.Name()); err == nil {
	// 					path = f.Name()
	// 					basename = filepath.Base(path)
	// 					location = fmt.Sprintf("file://%s", path)

	// 					defer f.Close()
	// 					fileinfo, _ := f.Stat()
	// 					size = fileinfo.Size()
	// 					h := sha1.New()
	// 					if _, err := io.Copy(h, f); err != nil {
	// 						log.Fatal(err)
	// 					}

	// 					checksum = fmt.Sprintf("sha1$%x", string(h.Sum(nil)))
	// 					if listingString != "" {
	// 						listingString = listingString + ","
	// 					}
	// 					listingString = listingString + "{\"checksum\": \"" + checksum + "\",\"basename\": \"" + basename + "\",\"location\": \"" + location + "\",\"path\": \"" + path + "\",\"class\": \"File\",\"size\": " + strconv.FormatInt(size, 10) + "}"
	// 				}
	// 			}
	// 			outputIdentifier := tool.Root.Outputs[0].ID
	// 			fmt.Println("{\"" + outputIdentifier + "\":{\"class\": \"Directory\",\"listing\": [" + listingString + "]}}")
	// 		}
	// 	}
	// }
	// // }}}

	return nil
}

// ensureArguments ...
func (tool *CommandLineTool) ensureArguments() []string {
	result := []string{}
	sort.Sort(tool.Root.Arguments)
	for i, arg := range tool.Root.Arguments {
		if arg.Binding != nil && arg.Binding.ValueFrom != nil {
			tool.Root.Arguments[i].Value = tool.AliasFor(arg.Binding.ValueFrom.Key())
		}
		result = append(result, tool.Root.Arguments[i].Flatten()...)
	}
	return result
}

// ensureInputs ...
func (tool *CommandLineTool) ensureInputs() (priors []string, result []string, err error) {
	defer func() {
		if i := recover(); i != nil {
			err = fmt.Errorf("failed to collate required inputs and provided params: %v", i)
		}
	}()
	sort.Sort(tool.Root.Inputs)
	for _, in := range tool.Root.Inputs {
		in, err = tool.ensureInput(in)
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
func (tool *CommandLineTool) ensureInput(input cwl.Input) (cwl.Input, error) {
	if provided, ok := tool.Parameters[input.ID]; ok {
		input.Provided = provided
	}
	if input.Default == nil && input.Binding == nil && input.Provided == nil {
		return input, fmt.Errorf("input `%s` doesn't have default field but not provided", input.ID)
	}
	if key, needed := input.Types[0].NeedRequirement(); needed {
		for _, req := range tool.Root.Requirements {
			for _, requiredtype := range req.Types {
				if requiredtype.Name == key {
					input.RequiredType = &requiredtype
					input.Requirements = tool.Root.Requirements
				}
			}
		}
	}
	return input, nil
}

// AliasFor ...
func (tool *CommandLineTool) AliasFor(key string) string {
	switch key {
	case "runtime.cores":
		return "2"
	}
	return ""
}

// generateCommand ...
func (tool *CommandLineTool) generateCommand(priors, arguments, inputs []string) (*exec.Cmd, error) {

	if len(tool.Root.BaseCommands) == 0 {
		return exec.Command("bash", "-c", tool.Root.Arguments[0].Binding.ValueFrom.Key()), nil
	}

	// Join all slices
	oneline := []string{}
	oneline = append(oneline, tool.Root.BaseCommands...)
	oneline = append(oneline, priors...)
	oneline = append(oneline, arguments...)
	oneline = append(oneline, inputs...)

	return exec.Command(oneline[0], oneline[1:]...), nil
}

// cmdが実行されるディレクトリを決める
// tool.Commandが生成されている必要がある
func (tool *CommandLineTool) cmdが実行されるディレクトリを決める() error {

	rootdir := filepath.Dir(tool.Root.Path)
	tool.Command.Dir = rootdir

	// If "Directory" is specified in "outputs", use it.
	for _, o := range tool.Root.Outputs {
		if o.Types[0].Type == "Directory" {
			tmpdir, err := ioutil.TempDir("/", "")
			if err != nil {
				return err
			}
			if err := os.Link(rootdir, tmpdir); err != nil {
				return err
			}
			tool.Command.Dir = tmpdir
		}
	}

	return nil
}

// 標準出力の行き先を決める
// tool.Command、およびtool.Command.Dirが先に決定されている必要がある
func (tool *CommandLineTool) 標準出力の行き先を決める() error {

	// Prefer "stdout" specified on ROOT
	if tool.Root.Stdout != "" {
		stdoutfilepath := filepath.Join(tool.Command.Dir, tool.Root.Stdout)
		stdoutfile, err := os.Create(stdoutfilepath)
		if err != nil {
			return err
		}
		tool.Command.Stdout = stdoutfile
		return nil
	}

	// Respect "stdout" specified in each "output"
	for _, o := range tool.Root.Outputs {
		if o.Types[0].Type == "stdout" {
			stdoutfilepath := filepath.Join(tool.Command.Dir, o.ID)
			stdoutfile, err := os.Create(stdoutfilepath)
			if err != nil {
				return err
			}
			tool.Command.Stdout = stdoutfile
		}
	}

	// nothing specified
	return nil
}

// 標準エラーの行き先を決める
func (tool *CommandLineTool) 標準エラーの行き先を決める() error {

	// Prefer "stderr" specified on ROOT
	if tool.Root.Stderr != "" {
		stderrfilepath := filepath.Join(tool.Command.Dir, tool.Root.Stderr)
		stderrfile, err := os.Create(stderrfilepath)
		if err != nil {
			return err
		}
		tool.Command.Stderr = stderrfile
		return nil
	}

	// Respect "stderr" specified in each "output"
	for _, o := range tool.Root.Outputs {
		if o.Types[0].Type == "stderr" {
			stderrfilepath := filepath.Join(tool.Command.Dir, o.ID)
			stderrfile, err := os.Create(stderrfilepath)
			if err != nil {
				return err
			}
			tool.Command.Stderr = stderrfile
		}
	}

	return nil
}

// outdirを決定する
func (tool *CommandLineTool) outdirを決定する() error {
	if tool.Outdir != "" {
		return nil
	}
	rootdir := filepath.Dir(tool.Root.Path)
	tool.Outdir = rootdir
	return nil
}

// outdirにいろいろ揃える。必要があればcwl.output.jsonをつくる
func (tool *CommandLineTool) outdirにいろいろ揃える() error {

	// If "cwl.output.json" exists on executed command directory,
	// dump the file contents on stdout.
	// This is described on official document.
	// See also https://www.commonwl.org/v1.0/CommandLineTool.html#Output_binding
	whatthefuck := filepath.Join(tool.Command.Dir, "cwl.output.json")
	if defaultout, err := os.Open(whatthefuck); err == nil {
		defer defaultout.Close()
		if _, err := io.Copy(os.Stdout, defaultout); err != nil {
			return err
		}
	}

	return nil
}

// Finalize closes all file desccriptors if needed.
func (tool *CommandLineTool) Finalize() error {
	return nil
}
