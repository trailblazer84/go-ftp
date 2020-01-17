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
	ftplog.Println("connect to ", server)

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
		for _, file := range c.Files {
			ftplog.Println("sending file -", file)
			fdesc, err := os.Open(file)
			if err != nil {
				ftplog.Println(err)
				return err
			}
			fst, _ := fdesc.Stat()
			wr, err := io.CopyN(rw, fdesc, fst.Size())
			if err != nil {
				ftplog.Println(err)
				return err
			}
			ftplog.Println("bytes sent -", wr)
		}
		break
	case "recv":
		for _, file := range c.Files {
			ftplog.Println("receiving file -", file)
			str, _ := rw.ReadString('\n')
			str = strings.Trim(str, "\n")
			sz, err := strconv.Atoi(str)
			if err != nil {
				ftplog.Println(str)
				return err
			}
			ftplog.Println("client/" + filepath.Base(file))
			fdesc, err := os.Create("client/" + filepath.Base(file))
			if err != nil {
				ftplog.Println(err)
				return err
			}
			wr, err := io.CopyN(fdesc, rw, int64(sz))
			if err != nil {
				return err
			}
			ftplog.Println("bytes received -", wr)
		}
		break
	}

	ftplog.Println("waiting to get result")
	str, err := rw.ReadString('\n')
	ftplog.Println(c.Cmd, " - ", str)
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
	ftplog.Println(c.Cmd, c.Files)
	switch c.Cmd {
	case "send":
		ftplog.Println("receiving file from ", conn.RemoteAddr())
		for i, file := range c.Files {
			ftplog.Println("server/" + filepath.Base(file))
			fdesc, err := os.Create("server/" + filepath.Base(file))
			if err != nil {
				str = "IO error\n"
				break
			}
			sz := c.Fsizes[i]
			wr, err := io.CopyN(fdesc, rw, sz)
			if err != nil {
				str = "IO error\n"
				break
			}
			ftplog.Println("bytes received -", wr)
		}
		break
	case "recv":
		ftplog.Println("sending file to", conn.RemoteAddr())
		for _, file := range c.Files {
			ftplog.Println(file)
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
			ftplog.Println("bytes sent -", wr)
		}
		break
	}

	if str == "" {
		str = "success\n"
	}

	ftplog.Println(str)
	rw.WriteString(str)
	rw.Flush()
}

func connectClient(c *ctx) error {

	ln, err := net.Listen("tcp", ":"+strconv.Itoa(mfport))
	if err != nil {
		return err
	}

	for {
		ftplog.Println("Waiting to accept connection from client")
		conn, err := ln.Accept()
		if err != nil {
			break
		}
		defer conn.Close()

		ftplog.Println("Accepted connection from client -", conn.RemoteAddr())

		go ftpio(conn)
	}

	return err
}
