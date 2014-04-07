package interpreter

import "os"
import "os/exec"
import "../parser"

type WhatToDoAfterCmd int

const (
	THROUGH  WhatToDoAfterCmd = 0
	CONTINUE WhatToDoAfterCmd = 1
	SHUTDOWN WhatToDoAfterCmd = 2
)

func Interpret(text string, hook func(cmd *exec.Cmd, IsBackground bool) (WhatToDoAfterCmd, error)) (WhatToDoAfterCmd, error) {
	statements := parser.Parse(text)
	for _, pipeline := range statements {
		var pipeIn *os.File = nil
		for _, state := range pipeline {
			//fmt.Println(state)
			cmd := new(exec.Cmd)
			cmd.Args = state.Argv
			cmd.Env = nil
			cmd.Dir = ""
			cmd.Stdout = os.Stdout
			cmd.Stdin = os.Stdin
			cmd.Stderr = os.Stderr
			if pipeIn != nil {
				cmd.Stdin = pipeIn
				pipeIn = nil
			}
			if state.Redirect[0].Path != "" {
				fd, err := os.Open(state.Redirect[0].Path)
				if err != nil {
					return CONTINUE, err
				}
				defer fd.Close()
				cmd.Stdin = fd
			}
			if state.Redirect[1].Path != "" {
				var fd *os.File
				var err error
				if state.Redirect[1].IsAppend {
					fd, err = os.OpenFile(state.Redirect[1].Path, os.O_APPEND, 0666)
				} else {
					fd, err = os.OpenFile(state.Redirect[1].Path, os.O_CREATE, 0666)
				}
				if err != nil {
					return CONTINUE, err
				}
				defer fd.Close()
				cmd.Stdout = fd
			}
			if state.Redirect[2].Path != "" {
				var fd *os.File
				var err error
				if state.Redirect[2].IsAppend {
					fd, err = os.OpenFile(state.Redirect[2].Path, os.O_APPEND, 0666)
				} else {
					fd, err = os.OpenFile(state.Redirect[2].Path, os.O_CREATE, 0666)
				}
				if err != nil {
					return CONTINUE, err
				}
				defer fd.Close()
				cmd.Stderr = fd
			}
			var err error = nil
			var pipeOut *os.File = nil
			if state.Term == "|" {
				pipeIn, pipeOut, err = os.Pipe()
				if err != nil {
					return CONTINUE, err
				}
				defer pipeIn.Close()
				cmd.Stdout = pipeOut
			}
			var whatToDo WhatToDoAfterCmd

			isBackGround := (state.Term == "|" || state.Term == "&")

			whatToDo, err = hook(cmd, isBackGround)
			if whatToDo == THROUGH {
				cmd.Path, err = exec.LookPath(state.Argv[0])
				if err == nil {
					if isBackGround {
						err = cmd.Start()
					} else {
						err = cmd.Run()
					}
				}
			}
			if pipeOut != nil {
				pipeOut.Close()
			}
			if whatToDo == SHUTDOWN {
				return SHUTDOWN, err
			}
			if err != nil {
				return CONTINUE, err
			}
		}
	}
	return CONTINUE, nil
}