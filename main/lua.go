package main

import (
	"fmt"
	"os"
	"runtime"
	"unsafe"

	"../completion"
	"../conio"
	"../dos"
	"../history"
	"../interpreter"
	"../lua"
	ole "../lua/ole"
)

const REGKEY_INTERPRETER = "nyagos.interpreter"

func setRegInt(L lua.Lua, it *interpreter.Interpreter) {
	L.PushValue(lua.LUA_REGISTRYINDEX)
	L.PushLightUserData(unsafe.Pointer(it))
	L.SetField(-2, REGKEY_INTERPRETER)
	L.Pop(1)
}

func getRegInt(L lua.Lua) *interpreter.Interpreter {
	L.PushValue(lua.LUA_REGISTRYINDEX)
	L.GetField(-1, REGKEY_INTERPRETER)
	rc := (*interpreter.Interpreter)(L.ToUserData(-1))
	L.Pop(2)
	return rc
}

func NyagosCallLua(L lua.Lua, it *interpreter.Interpreter, nargs int, nresult int) error {
	save := getRegInt(L)
	setRegInt(L, it)
	err := L.Call(nargs, nresult)
	setRegInt(L, save)
	return err
}

var orgArgHook func(*interpreter.Interpreter, []string) ([]string, error)

var luaArgsFilter lua.Pushable = lua.TNil{}

func newArgHook(it *interpreter.Interpreter, args []string) ([]string, error) {
	L := NewNyagosLua()
	defer L.Close()
	L.Push(luaArgsFilter)
	if !L.IsFunction(-1) {
		return orgArgHook(it, args)
	}
	L.NewTable()
	for i := 0; i < len(args); i++ {
		L.PushString(args[i])
		L.RawSetI(-2, lua.Integer(i))
	}
	if err := L.Call(1, 1); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return orgArgHook(it, args)
	}
	if L.GetType(-1) != lua.LUA_TTABLE {
		return orgArgHook(it, args)
	}
	newargs := []string{}
	for i := lua.Integer(0); true; i++ {
		L.PushInteger(i)
		L.GetTable(-2)
		if L.GetType(-1) == lua.LUA_TNIL {
			break
		}
		arg1, arg1err := L.ToString(-1)
		if arg1err == nil {
			newargs = append(newargs, arg1)
		} else {
			fmt.Fprintln(os.Stderr, arg1err.Error())
		}
		L.Pop(1)
	}
	return orgArgHook(it, newargs)
}

var orgOnCommandNotFound func(*interpreter.Interpreter, error) error

var luaOnCommandNotFound lua.Pushable = lua.TNil{}

func on_command_not_found(inte *interpreter.Interpreter, err error) error {
	L := NewNyagosLua()
	defer L.Close()

	L.Push(luaOnCommandNotFound)
	if !L.IsFunction(-1) {
		L.Pop(1)
		return orgOnCommandNotFound(inte, err)
	}
	L.NewTable()
	for key, val := range inte.Args {
		L.PushString(val)
		L.RawSetI(-2, lua.Integer(key))
	}
	err1 := L.Call(1, 1)
	defer L.Pop(1)
	if err1 != nil {
		return err
	}
	if L.ToBool(-1) {
		return nil
	} else {
		return orgOnCommandNotFound(inte, err)
	}
}

func emptyToNil(s string) lua.Pushable {
	if s == "" {
		return &lua.TNil{}
	} else {
		return &lua.TString{s}
	}
}

var nyagos_table_member map[string]lua.Pushable

func getNyagosTable(L lua.Lua) int {
	index, index_err := L.ToString(2)
	if index_err != nil {
		return L.Push(nil, index_err.Error())
	}
	if entry, entry_ok := nyagos_table_member[index]; entry_ok {
		return L.Push(entry)
	} else if index == "exe" {
		if exeName, exeNameErr := dos.GetModuleFileName(); exeNameErr != nil {
			return L.Push(nil, exeNameErr.Error())
		} else {
			L.PushString(exeName)
			return 1
		}
	} else {
		L.PushNil()
		return 1
	}
}

type IProperty interface {
	Push(lua.Lua) int
	Set(lua.Lua, int) error
}

func setNyagosTable(L lua.Lua) int {
	index, index_err := L.ToString(2)
	if index_err != nil {
		return L.Push(nil, index_err)
	}
	if current_value, exists := nyagos_table_member[index]; exists {
		if property, castOk := current_value.(IProperty); castOk {
			if err := property.Set(L, 3); err != nil {
				return L.Push(nil, err)
			} else {
				return L.Push(true)
			}
		} else {
			value, value_err := L.ToPushable(3)
			if value_err != nil {
				return L.Push(nil, value_err)
			}
			nyagos_table_member[index] = value
			return L.Push(true)
		}
	} else {
		fmt.Fprintf(os.Stderr, "nyagos.%s: reserved variable.\n", index)
		return L.Push(nil)
	}
}

var share_table = map[string]lua.Pushable{}

func getShareTable(L lua.Lua) int {
	key, keyErr := L.ToString(-1)
	if keyErr != nil {
		return L.Push(nil, keyErr)
	}
	if value, ok := share_table[key]; ok {
		return L.Push(value)
	} else {
		L.PushNil()
		return 1
	}
}

func setShareTable(L lua.Lua) int {
	key, keyErr := L.ToString(-2)
	if keyErr != nil {
		return L.Push(nil, keyErr)
	}
	value, valErr := L.ToPushable(-1)
	if valErr != nil {
		return L.Push(nil, valErr)
	}
	share_table[key] = value
	return 1
}

var hook_setuped = false

func NewNyagosLua() lua.Lua {
	this := lua.New()
	this.OpenLibs()

	this.Push(lua.NewVirtualTable(getNyagosTable, setNyagosTable))
	this.SetGlobal("nyagos")

	this.Push(lua.NewVirtualTable(getShareTable, setShareTable))
	this.SetGlobal("share")

	if !hook_setuped {
		orgArgHook = interpreter.SetArgsHook(newArgHook)

		orgOnCommandNotFound = interpreter.OnCommandNotFound
		interpreter.OnCommandNotFound = on_command_not_found
		hook_setuped = true
	}
	return this
}

var silentmode = false

func init() {
	nyagos_table_member = map[string]lua.Pushable{
		"access":               &lua.TGoFunction{cmdAccess},
		"alias":                lua.NewVirtualTable(cmdGetAlias, cmdSetAlias),
		"antihistquot":         lua.StringProperty{&history.DisableMarks},
		"atou":                 &lua.TGoFunction{cmdAtoU},
		"bindkey":              &lua.TGoFunction{cmdBindKey},
		"chdir":                &lua.TGoFunction{cmdChdir},
		"create_object":        &lua.TGoFunction{ole.CreateObject},
		"commit":               emptyToNil(commit),
		"commonprefix":         &lua.TGoFunction{cmdCommonPrefix},
		"default_prompt":       &lua.TGoFunction{nyagosPrompt},
		"env":                  lua.NewVirtualTable(cmdGetEnv, cmdSetEnv),
		"eval":                 &lua.TGoFunction{cmdEval},
		"exec":                 &lua.TGoFunction{cmdExec},
		"histchar":             lua.StringProperty{&history.Mark},
		"getalias":             &lua.TGoFunction{cmdGetAlias},
		"getenv":               &lua.TGoFunction{cmdGetEnv},
		"gethistory":           &lua.TGoFunction{cmdGetHistory},
		"getkey":               &lua.TGoFunction{cmdGetKey},
		"getviewwidth":         &lua.TGoFunction{cmdGetViewWidth},
		"getwd":                &lua.TGoFunction{cmdGetwd},
		"glob":                 &lua.TGoFunction{cmdGlob},
		"lines":                &lua.TGoFunction{cmdLines},
		"loadfile":             &lua.TGoFunction{cmdLoadFile},
		"netdrivetounc":        &lua.TGoFunction{cmdNetDriveToUNC},
		"open":                 &lua.TGoFunction{cmdOpenFile},
		"pathjoin":             &lua.TGoFunction{cmdPathJoin},
		"quotation":            lua.StringProperty{&conio.Delimiters},
		"raweval":              &lua.TGoFunction{cmdRawEval},
		"rawexec":              &lua.TGoFunction{cmdRawExec},
		"resetcharwidth":       &lua.TGoFunction{cmdResetCharWidth},
		"setalias":             &lua.TGoFunction{cmdSetAlias},
		"setenv":               &lua.TGoFunction{cmdSetEnv},
		"setrunewidth":         &lua.TGoFunction{cmdSetRuneWidth},
		"shellexecute":         &lua.TGoFunction{cmdShellExecute},
		"silentmode":           &lua.BoolProperty{&silentmode},
		"stat":                 &lua.TGoFunction{cmdStat},
		"stamp":                emptyToNil(stamp),
		"utoa":                 &lua.TGoFunction{cmdUtoA},
		"which":                &lua.TGoFunction{cmdWhich},
		"write":                &lua.TGoFunction{cmdWrite},
		"writerr":              &lua.TGoFunction{cmdWriteErr},
		"goarch":               &lua.TString{runtime.GOARCH},
		"goversion":            &lua.TString{runtime.Version()},
		"version":              emptyToNil(version),
		"prompt":               lua.Property{&prompt_hook},
		"argsfilter":           lua.Property{&luaArgsFilter},
		"filter":               lua.Property{&luaFilter},
		"on_command_not_found": lua.Property{&luaOnCommandNotFound},
		"completion_hook":      lua.Property{&completion.Hook},
	}
}
