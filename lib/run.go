package lib

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/fatih/color"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

var model *MkModel

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func ResolveConfig() (err error) {
	if CLI.File != "" {
		if CLI.Init && FileExists(CLI.File) {
			println(fmt.Sprintf("Using file: %s", CLI.File))
			return
		} else {
			err = errors.New(fmt.Sprintf("Config File %s not found", CLI.File))
		}
	}
	deffiles := []string{".mk.yaml", ".mk", "mk", "mk.yaml"}
	for _, v := range deffiles {
		if CLI.Init || FileExists(v) {
			println(fmt.Sprintf("Using file: %s", v))
			CLI.File = v
			return
		}
	}
	err = errors.New(fmt.Sprintf("No Config File found"))
	return
}

func ResolveTask(n string) *MkTask {
	osname := runtime.GOOS
	arch := runtime.GOARCH
	vars["BASETASK"] = n
	t, ok := model.Tasks[n+"_"+osname+"_"+arch]
	if ok {
		vars["TASK"] = n + "_" + osname + "_" + arch
		return t
	}
	t, ok = model.Tasks[n+"_"+osname]
	if ok {
		vars["TASK"] = n + "_" + osname
		return t
	}
	vars["TASK"] = n

	return model.Tasks[n]
}

func Log(i int, task string, stream string, b string) {
	timeLb := time.Now().Format("15:04:05")
	tab := ""
	for x := 0; x < i; x++ {
		tab = tab + "\t"
	}
	if stream == "O" {
		println(color.GreenString(fmt.Sprintf("%s %s [%s - %s]: %s", timeLb, tab, task, stream, b)))
	} else if stream == "E" {
		println(color.RedString(fmt.Sprintf("%s %s [%s - %s]: %s", timeLb, tab, task, stream, b)))
	} else {
		println(fmt.Sprintf("%s %s [%s - %s]: %s", timeLb, tab, task, stream, b))
	}

}

func RunTask(name string, t *MkTask, l int) error {

	for k, v := range t.Env {
		env[k] = v
	}

	for k, v := range t.Vars {
		vars[k] = v
	}

	Log(l, name, "**", fmt.Sprintf("Starting %s", name))
	if len(t.Pre) > 0 {
		Log(l, name, "**", fmt.Sprintf("Will run Pre tasks: [%s]", strings.Join(t.Pre, ",")))
		for _, v := range t.Pre {
			caller, repeat := model.Stack[v]
			if repeat {
				Log(l, name, "**", fmt.Sprintf("Task  %s already called by %s - skipping.", v, caller))
				continue
			} else {
				model.Stack[v] = name
			}

			pr, ok := model.Tasks[v]
			if ok {
				err := RunTask(v, pr, l+1)
				if err != nil {
					return err
				}
			} else {
				return errors.New(fmt.Sprintf("Task %s, prereq of: %s not found in model", v, name))
			}
		}
	}
	var c *exec.Cmd
	if runtime.GOOS == "windows" {
		c = exec.Command("cmd.exe")
	} else {
		c = exec.Command("sh")
	}
	for k, v := range env {
		c.Env = append(c.Env, fmt.Sprintf("%s=%s", k, v))
	}

	pi, _ := c.StdinPipe()

	po, err := c.StdoutPipe()

	if err != nil {
		log.Printf(err.Error())
		return err
	}
	scannero := bufio.NewScanner(po)

	pe, err := c.StderrPipe()
	if err != nil {
		log.Printf(err.Error())
		return err
	}
	scannere := bufio.NewScanner(pe)

	wg := sync.WaitGroup{}

	wg.Add(2)

	ch := make(chan string)

	go func() {
		time.Sleep(time.Millisecond * 100)
		defer wg.Done()
		for scannero.Scan() {
			line := scannero.Text()
			if strings.Contains(line, "__CMD_ENDED__:") {
				ret := strings.Split(line, "__CMD_ENDED__:")[1]
				if ret != "%errorlevel%" {
					ch <- ret
				}
			} else {
				Log(l+1, name, "O", line)
			}
		}

	}()

	go func() {
		time.Sleep(time.Millisecond * 100)
		defer wg.Done()
		for scannere.Scan() {
			Log(l+1, name, "E", scannere.Text())
		}
	}()

	err = c.Start()

	if err != nil {
		return err
	}

	lines := bufio.NewScanner(strings.NewReader(t.Cmd))

	//if runtime.GOOS == "windows" {
	//	pi.Write([]byte("@echo off\n"))
	//}

	for lines.Scan() {

		txt := lines.Text()

		for k, v := range vars {
			txt = strings.Replace(txt, fmt.Sprintf("${%s}", k), v, -1)
		}
		if strings.HasPrefix(txt, "@") {
			txt = strings.TrimLeft(txt, "@")
		} else {
			Log(l+1, name, "I", txt)
		}

		if strings.HasPrefix(txt, "mk:") {
			txt = strings.Replace(txt, "mk:", "", 1)
			txt = strings.TrimSpace(txt)
			err = RunMkCmd(l+1, txt)
			if err != nil {
				log.Printf("Error running: %s", txt)
			}
		} else {

			_, err = pi.Write([]byte(txt + "\n"))

			if runtime.GOOS == "windows" {
				_, err = pi.Write([]byte("echo __CMD_ENDED__:%errorlevel%" + "\n"))
			} else {
				_, err = pi.Write([]byte("echo __CMD_ENDED__:$?" + "\n"))
			}

			ret := <-ch
			if ret != "0" && t.Onerror == "skip" {
				Log(l+1, name, "E", "Error Code: "+ret+" will continue, but watch out")

			} else if ret != "0" && t.Onerror != "skip" {
				Log(l+1, name, "E", "Error Code is: "+ret+" ABORTING")
				break
			}
		}

	}
	_, err = pi.Write([]byte("exit\n"))
	if err != nil {
		return err
	}

	wg.Wait()
	err = c.Wait()
	Log(l, name, "**", "End")
	return err

}

func DumpEnv() {
	if CLI.Env {
		println("Vars:")

		varnames := make([]string, 0)
		for k := range vars {
			varnames = append(varnames, k)
		}
		sort.Strings(varnames)
		for _, k := range varnames {
			v := vars[k]
			println(fmt.Sprintf("%s => %s", k, v))
		}

		println("========")
		println("Env:")

		envnames := make([]string, 0)
		for k := range env {
			envnames = append(envnames, k)
		}
		sort.Strings(envnames)
		for _, k := range envnames {
			v := env[k]
			println(fmt.Sprintf("%s => %s", k, v))
		}

		os.Exit(0)
	}
}

func Prepare() error {

	InitCli()
	err := ResolveConfig()
	if err != nil {
		return err
	}

	if CLI.Init {
		err = InitFile()
		if err != nil {
			return err
		}
		os.Exit(0)
	}

	if CLI.DumpValidator {
		err = DumpValidator()
		if err != nil {
			log.Printf("Error DumpValidator: %s", err.Error())
		}
		os.Exit(0)
	}

	bs, err := os.ReadFile(CLI.File)
	if err != nil {
		panic(err)
	}
	model = &MkModel{}
	err = yaml.Unmarshal(bs, model)
	if err != nil {
		return err
	}
	if model.Default == "" {
		model.Default = "main"
	}

	for k, v := range model.Tasks {
		v.Name = k
	}

	if CLI.List {
		println("Tasks:")
		tasknames := make([]string, 0)
		for k := range model.Tasks {
			tasknames = append(tasknames, k)
		}
		sort.Strings(tasknames)
		for _, k := range tasknames {
			v := model.Tasks[k]
			def := ""
			if v.Name == model.Default {
				def = "DEF>"
			} else {
				def = ""
			}
			println(fmt.Sprintf("%s	%s [%s]: %s", def, k, strings.Join(v.Pre, ","), v.Help))
		}
		os.Exit(0)
	}

	if len(CLI.Tasks) < 1 || CLI.Tasks[0] == "" || CLI.Tasks[0] == "." {
		CLI.Tasks = []string{model.Default}
	}

	return err
}

var env map[string]string
var vars map[string]string

func Run() error {
	env = make(map[string]string)
	vars = make(map[string]string)
	envstrs := os.Environ()
	for _, k := range envstrs {
		parts := strings.Split(k, "=")
		ek := strings.TrimSpace(parts[0])
		ev := os.Getenv(ek)
		env[ek] = ev
	}

	err := Prepare()
	if err != nil {
		log.Printf("Could not execute: %s", err.Error())
		os.Exit(0)
	}

	for k, v := range model.Env {
		env[k] = v
	}

	for k, v := range model.Vars {
		vars[k] = v
	}
	now := time.Now()
	vars["DT_YYMMDDHHmmss"] = now.Format("060102150405")
	vars["DT_YYMMDD"] = now.Format("060102")
	vars["DS_TS"] = fmt.Sprintf("%d", now.Unix())

	u, err := user.Current()
	if err != nil {
		panic(err)
	}
	vars["USERNAME"] = u.Username
	vars["HOMEDIR"] = u.HomeDir

	if err != nil {
		return err
	}

	DumpEnv()
	model.Stack = make(map[string]string)
	for _, v := range CLI.Tasks {
		tasko := ResolveTask(v)
		if tasko != nil {
			err = RunTask(v, tasko, 0)
		} else {
			return errors.New(fmt.Sprintf("No task named %s found", v))
		}
	}

	return err
}
