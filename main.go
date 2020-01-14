package main

import (
	"flag"
	"fmt"
	"github.com/google/goterm/term"
	"net"
	"os"
	"path/filepath"
	"strings"
)

type ctx struct {
	client bool
	sip    string
	Cmd    string
	Files  []string
	Fsizes []int64
}

var (
	sip    = flag.String("sip", "", "server IP")
	cmd    = flag.String("cmd", "", "command to use (send / recv)")
	files  = flag.String("files", "", "list of files (',' separated")
	client = flag.Bool("client", false, "run as a client")
)

func removeDupFiles(f []string) []string {
	keys := make(map[string]bool)
	files := []string{}

	for _, entry := range f {
		v := keys[entry]
		if v == false {
			keys[entry] = true
			files = append(files, entry)
		}
	}

	return files
}

func validateFiles(c *ctx) bool {
	dataFiles := strings.Split(*files, ",")
	for _, file := range dataFiles {
		absPath, err := filepath.Abs(file)
		if err != nil {
			fmt.Println("error:", err)
			return false
		}

		f, err := filepath.Glob(absPath)
		if err != nil {
			fmt.Println("error:", err)
			return false
		}

		if len(f) != 0 {
			c.Files = append(c.Files, f...)
		} else {
			if strings.Compare(strings.ToUpper(*cmd), "SEND") == 0 {
				_, err := os.Stat(absPath)
				if os.IsNotExist(err) {
					fmt.Println(file, "- does not exist!")
					return false
				}
			}
			c.Files = append(c.Files, absPath)
		}
	}

	c.Files = removeDupFiles(c.Files)
	if strings.Compare(strings.ToUpper(*cmd), "SEND") == 0 {
		for _, file := range c.Files {
			fst, _ := os.Stat(file)
			fmt.Println(file, fst.Size())
			c.Fsizes = append(c.Fsizes, fst.Size())
		}
	}
	return true
}

func validateIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

func validateArgs(c *ctx) bool {
	if *client == true {
		if validateIP(*sip) == false {
			fmt.Println("Invalid server ip!")
			return false
		}
		if (strings.Compare(strings.ToUpper(*cmd), "SEND") != 0) &&
			(strings.Compare(strings.ToUpper(*cmd), "RECV") != 0) {
			fmt.Println("Invalid command -", *cmd, "!")
			return false
		}
		if *files == "" {
			fmt.Println("Invalid files!")
			return false
		}
		if validateFiles(c) == false {
			fmt.Println("Invalid files!")
			return false
		}
	}

	return true
}

func handleClient(c *ctx) {
	fmt.Println(term.Redf("my_ftp client"))
	fmt.Println("server IP -", c.sip)
	fmt.Println("cmd -", c.Cmd)
	fmt.Println("files -")
	for _, file := range c.Files {
		fmt.Println(file)
	}

	err := connectServer(c)
	if err != nil {
		fmt.Println("connnection failed -", err)
	}
}

func handleServer(c *ctx) {
	fmt.Println(term.Redf("my_ftp server"))
	err := connectClient(c)
	if err != nil {
		fmt.Println("connnection failed -", err)
	}
}

func main() {
	flag.Parse()
	ctx := ctx{client: *client, sip: *sip, Cmd: *cmd}

	if validateArgs(&ctx) == false {
		fmt.Println("usage ->")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if ctx.client {
		handleClient(&ctx)
	} else {
		handleServer(&ctx)
	}
}
