TCP server
==========

Basic API for TCP server. Inspired by the [article](https://sahilm.com/tcp-servers-that-run-like-clockwork/).
The main idea is to move low-level work with TCP sockets to API level and provide a possibility to you to reduce amount of code. 

## Usage

Just make a simple type, that implements Handler interface. It should contain only one method:
```go
func Handle(data []byte) ([]byte, err)
```
, see [example](example/main.go).

and execute it like:
```bash
$ go run example/main.go
$ telnet 0.0.0.0 9000
Trying 0.0.0.0...
Connected to 0.0.0.0.
Escape character is '^]'.
hi
you said: hi
bye
you said: bye
# waiting here for timeout
Connection closed by foreign host.
```
