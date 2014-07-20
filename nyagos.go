package main

import "bufio"
import "fmt"
import "os"
import "os/signal"
import "path/filepath"
import "regexp"
import "strings"

import "github.com/mattn/go-runewidth"
import "github.com/shiena/ansicolor"

import "./completion"
import "./conio"
import "./exename"
import "./history"
import "./interpreter"
import "./lua"
import "./option"
import "./prompt"

func signalOff() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	go func() {
		for _ = range ch {

		}
	}()
}

var rxAnsiEscCode = regexp.MustCompile("\x1b[^a-zA-Z]*[a-zA-Z]")

func main() {
	signalOff()

	// KeyBind += completion Module
	conio.KeyMap['\t'] = completion.KeyFuncCompletion
	conio.ZeroMap[conio.K_UP] = history.KeyFuncHistoryUp
	conio.ZeroMap[conio.K_DOWN] = history.KeyFuncHistoryDown
	conio.KeyMap['P'&0x1F] = history.KeyFuncHistoryUp
	conio.KeyMap['N'&0x1F] = history.KeyFuncHistoryDown

	// ANSI Escape Sequence Support
	ansiOut := ansicolor.NewAnsiColorWriter(os.Stdout)

	// Lua extension
	L := lua.NewLua()
	L.OpenLibs()
	lua.SetFunctions(L)
	defer L.Close()

	// Parameter Parsing
	argc := 0
	option.Parse(func() (string, bool) {
		argc++
		if argc < len(os.Args) {
			return os.Args[argc], true
		} else {
			return "", false
		}
	})

	appData := filepath.Join(os.Getenv("APPDATA"), "NYAOS_ORG")
	os.Mkdir(appData, 0777)
	histPath := filepath.Join(appData, "nyagos.history")
	history.Load(histPath)
	defer history.Save(histPath)

	exeName, exeNameErr := exename.Query()
	if exeNameErr != nil {
		fmt.Fprintln(os.Stderr, exeNameErr)
	}
	exeFolder := filepath.Dir(exeName)
	nyagos_lua := filepath.Join(exeFolder, "nyagos.lua")
	if _, err := os.Stat(nyagos_lua); err == nil {
		L.Call(nyagos_lua)
	}
	rcPath := filepath.Join(exeFolder, "nyagos.rc")
	if fd, err := os.Open(rcPath); err == nil {
		defer fd.Close()
		scr := bufio.NewScanner(fd)
		for scr.Scan() {
			interpreter.Interpret(scr.Text(), option.CommandHooks, nil)
		}
	}

	for {
		line, cont := conio.ReadLine(
			func() int {
				text := prompt.Format2Prompt(os.Getenv("PROMPT"))
				fmt.Fprint(ansiOut, text)
				text = rxAnsiEscCode.ReplaceAllString(text, "")
				lfPos := strings.LastIndex(text, "\n")
				if lfPos >= 0 {
					text = text[lfPos+1:]
				}
				return runewidth.StringWidth(text)
			})
		if cont == conio.ABORT {
			break
		}
		var isReplaced bool
		line, isReplaced = history.Replace(line)
		if isReplaced {
			fmt.Fprintln(os.Stdout, line)
		}
		history.Push(line)
		whatToDo, err := interpreter.Interpret(line, option.CommandHooks, nil)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		if whatToDo == interpreter.SHUTDOWN {
			break
		}
	}
}
