package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"strings"
)

var tok_delims = []string{"|", ">", "<"}

func main() {
	inputchan := make(chan [][]string, 1)
	username := getUsername()
	hostname, _ := os.Hostname()

	for {
		goshInput(username, hostname, inputchan)
		goshExec(inputchan)
	}
}

func goshInput(username, hostname string, inputchan chan [][]string) {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	pfstr := "/home/" + username + "/"
	if strings.HasPrefix(wd, pfstr) {
		wd = strings.Replace(wd, pfstr, "~/", 1)
	}

	pstr := (username + "@" + hostname + ":" + wd + "$ ")
	rl := bufio.NewReader(os.Stdin)

	fmt.Print(pstr)
	line, _, err := rl.ReadLine()
	if err != nil { // io.EOF
		fmt.Println(err)
		switch err {
		case io.EOF:
			{
				os.Exit(0)
			}
		}
	}

	args := strings.Split(string(line), "|")
	cmds := [][]string{}

	for _, arg := range args {
		cmd := strings.Split(strings.TrimSpace(arg), " ")
		cmds = append(cmds, cmd)
	}

	inputchan <- cmds
}

func goshExec(inputchan chan [][]string) {
	cmdargs := <-inputchan
	args := ""
	cmd := ""
	if len(cmdargs) == 1 {
		first_cmd := cmdargs[0]
		cmd = first_cmd[0]
		if len(first_cmd) > 1 {
			args = cmdargs[0][1]
		} else {
			args = ""
		}

		if cmd == "" {
			return
		} else if cmd == "cd" {
			os.Chdir(args)
			return
		} else if cmd == "exit" {
			os.Exit(0)
			return
		}
	}

	cmds := []*exec.Cmd{}

	for _, cmdarg := range cmdargs {
		cmd := cmdarg[0]
		args := cmdarg[1:]

		ecmd := exec.Command(cmd, args...)
		cmds = append(cmds, ecmd)
	}

	for i := 0; i <= len(cmds)-1; i++ {
		if i == 0 {
			cmds[i].Stdin = os.Stdin
			//cmds[i].Stdout = os.Stdout
			//cmds[i].Stderr = os.Stderr
		} else {
			cmds[i].Stdin, _ = cmds[i-1].StdoutPipe()
		}
	}

	cmds[len(cmds)-1].Stdout = os.Stdout

	for _, cmd := range cmds {
		err := cmd.Start()
		if err != nil {
			fmt.Println(err)
		}
	}
	for _, cmd := range cmds {
		err := cmd.Wait()
		if err != nil {
			fmt.Println(err)
		}
	}
}
func getUsername() string {
	username := ""

	user, err := user.Current()
	if err != nil {
		panic(err)
	} else {
		username = user.Username
	}

	return username
}
