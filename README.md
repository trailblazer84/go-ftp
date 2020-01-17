# go-ftp
Experiments with Golang. Simple file transfer client server program

```
usage ->
  -client
    	run as a client
  -cmd string
    	command to use (send / recv)
  -debug
    	add debugging
  -files string
    	list of files (',' separated
  -sip string
    	server IP
  -debug
      enable debugging
```

**EXAMPLE**

**server**
`go run main.go ftp_connection.go`

**client**
*send*
`go run main.go ftp_connection.go -client -sip 127.0.0.1 -files client/test.txt -cmd send`
*file from **client/test.txt** on client system gets received in **server/** directory on the server system*

*recv*
`go run main.go ftp_connection.go -client -sip 127.0.0.1 -files server/test.txt -cmd recv`
*file from **server/test.txt** on server system gets received in **client/** directory on the client system*

