package main

import (
	"flag"
	"fmt"
	"log"
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
	sip     = flag.String("sip", "", "server IP")
	cmd     = flag.String("cmd", "", "command to use (send / recv)")
	files   = flag.String("files", "", "list of files (',' separated")
	client  = flag.Bool("client", false, "run as a client")
	debug   = flag.Bool("debug", false, "add debugging")
	logFile = "go-ftp.log"
)

var ftplog *log.Logger

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

	if *debug == true {
		lfname := "server/" + logFile
		if *client == true {
			lfname = "client/" + logFile
		}

		logFD, err := os.Create(lfname)
		if err != nil {
			fmt.Println(logFile, "creation failed!")
			return false
		}
		ftplog = log.New(logFD, "go-ftp", log.Lshortfile)
		fmt.Println("logging enabled - ", lfname)
	}

	return true
}

func handleClient(c *ctx) {
	ftplog.Println("my_ftp client")
	ftplog.Println("server IP -", c.sip)
	ftplog.Println("cmd -", c.Cmd)
	ftplog.Println("files -")
	for _, file := range c.Files {
		ftplog.Println(file)
	}

	err := connectServer(c)
	if err != nil {
		fmt.Println("failed!")
	}
}

func handleServer(c *ctx) {
	ftplog.Println("my_ftp server")
	err := connectClient(c)
	if err != nil {
		fmt.Println("failed!")
	}
}

func main() {

	os.Mkdir("client", os.ModeDir|os.ModePerm)
	os.Mkdir("server", os.ModeDir|os.ModePerm)

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
