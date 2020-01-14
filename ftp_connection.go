package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	// mfport = TCP port number
	mfport int = 8787
	// cto = 2 second timeout
	cto int = 2
)

func connectServer(c *ctx) error {
	server := c.sip + ":" + strconv.Itoa(mfport)
	fmt.Println("connect to ", server)

	conn, err := net.DialTimeout("tcp", server, time.Duration(cto)*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	enc := gob.NewEncoder(rw)
	err = enc.Encode(c)
	if err != nil {
		return err
	}
	rw.Flush()

	switch c.Cmd {
	case "send":
		fmt.Println("sending file -")
		for _, file := range c.Files {
			fmt.Println(file)
			fdesc, err := os.Open(file)
			if err != nil {
				return err
			}
			fst, _ := fdesc.Stat()
			wr, err := io.CopyN(rw, fdesc, fst.Size())
			if err != nil {
				return err
			}
			fmt.Println("bytes sent -", wr)
		}
		break
	case "recv":
		fmt.Println("receiving file -")
		for _, file := range c.Files {
			fmt.Println("client/" + filepath.Base(file))
			fdesc, err := os.Create("client/" + filepath.Base(file))
			if err != nil {
				return err
			}
			str, _ := rw.ReadString('\n')
			str = strings.Trim(str, "\n")
			sz, err := strconv.Atoi(str)
			if err != nil {
				return err
			}
			wr, err := io.CopyN(fdesc, rw, int64(sz))
			if err != nil {
				return err
			}
			fmt.Println("bytes received -", wr)
		}
		break
	}

	fmt.Println("waiting to get result")
	str, err := rw.ReadString('\n')
	fmt.Println(c.Cmd, " - ", str)

	return nil
}

func ftpio(conn net.Conn) {
	var c ctx
	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	enc := gob.NewDecoder(rw)
	err := enc.Decode(&c)
	if err != nil {
		fmt.Println(err)
	}

	str := ""
	fmt.Println(c.Cmd, c.Files)
	switch c.Cmd {
	case "send":
		fmt.Println("receiving file from ", conn.RemoteAddr())
		for i, file := range c.Files {
			fmt.Println("server/" + filepath.Base(file))
			fdesc, err := os.Create("server/" + filepath.Base(file))
			if err != nil {
				str = "does not exist\n"
				break
			}
			sz := c.Fsizes[i]
			wr, err := io.CopyN(fdesc, rw, sz)
			if err != nil {
				str = "IO error\n"
				break
			}
			fmt.Println("bytes received -", wr)
		}
		break
	case "recv":
		fmt.Println("sending file to", conn.RemoteAddr())
		for _, file := range c.Files {
			fmt.Println(file)
			fst, err := os.Stat(file)
			if os.IsNotExist(err) {
				str = "does not exist\n"
				break
			}
			sz := int(fst.Size())
			szstr := strconv.Itoa(sz)
			rw.WriteString(szstr + "\n")
			rw.Flush()

			fdesc, _ := os.Open(file)
			wr, err := io.CopyN(rw, fdesc, fst.Size())
			if err != nil {
				str = "IO error\n"
				break
			}
			fmt.Println("bytes sent -", wr)
		}
		break
	}

	if str == "" {
		str = "success\n"
	}

	fmt.Println(str)
	rw.WriteString(str)
	rw.Flush()
}

func connectClient(c *ctx) error {

	ln, err := net.Listen("tcp", ":"+strconv.Itoa(mfport))
	if err != nil {
		return err
	}

	for {
		fmt.Println("Waiting to accept connection from client")
		conn, err := ln.Accept()
		if err != nil {
			break
		}
		defer conn.Close()

		fmt.Println("Accepted connection from client -", conn.RemoteAddr())

		go ftpio(conn)
	}

	return err
}
