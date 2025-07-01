# Distributed Key value store 

### Setup Instructions 

Build: `go build -o kvstore .`

Terminal 1:
```sh
./kvstore --node-id node1 --http-port 2222 --raft-port 8222
```

Terminal 2:
```sh
./kvstore --node-id node1 --http-port 2223 --raft-port 8223
```

Terminal 3:
```sh
curl "http://localhost:2222/Join?followerId=node2followerAddr=localhost:8223"
```

Everything is setup. Now you can start pushing data to the leader node and pulling data from any node like so:

Terminal 3:
```sh
curl "localhost:2222/Set/hello/there"
```

and

```sh
curl "localhost:2222/Set/hello/there"
```

