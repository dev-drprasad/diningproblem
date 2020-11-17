👉🏼👉🏼👉🏼 PUT THIS DIRECTOR IN GOPATH. (`~/go/src/diningproblem`) 👈🏼👈🏼👈🏼

This solution contains three files namely

1.  philosopher.go
2.  fork.go
3.  stalker.go

### philosopher.go

Arguments for `philosopher.go` are

1.  `name` : name of philosopher. will be used to register with printer process
2.  `listen`: this is address of philosopher where they listen for messages. form `<ip>:<port>`
3.  `neighbour`: listening address of next philosopher. form `<ip>:<port>`. Ex: "0.0.0.0:6002" , ":6002"
4.  `fork`: address of left fork. form `<ip>:<port>`
5.  `fork2`: address of left fork. form `<ip>:<port>`
6.  `the-initiator`: If passed, philosopher will act as leader.

### fork.go

Arguments for `fork.go` are

1.  `listen`: this is address of philosopher where they listen for messages. form `<ip>:<port>`

### stalker.go

Arguments for `stalker.go` are

1.  `listen`: this is address of philosopher where they listen for messages. form `<ip>:<port>`

### Coordination Process

1. All philosophers start waiting for neighbours indefinitely until they reach.
2. Leader (`--the-initiator`) will keep on ask its next neighbour a question : `ALL JOINED ?`. neighbour will pass it down same message to its neighbour if available. When this message reaches leader, all philosophers joined. Then the leader will send new message `LET's EAT` to next neighbor and neighbour passed it next neighbor. when this message reaches to philosopher, leader initiates dining process.

### Dining Process

Once a philosopher gets message `LET's EAT`, it starts eating process.

1. Initially philosopher starts with **thinking** state.
2. Once **thinking** is done, its go to **waiting** state means it tries to get forks
3. If it gets two forks, it starts **eating** process, or else it give up forks and goes to **thinking** and retries again
4. Once **eating** done, it returns forks
5. This cycle will repeat. no of cycles can be changed using variable `cycles` in `philosopher.go`

**eating**, **thinking** times are random. can be configured using variable `laziness` in `philosopher.go`

### Get Forks Process

When philosopher goes to **waiting** state,

1. it makes request to right left (the `--fork` one).
2. If right fork is available fork returns saying **OK**, if its not available, fork wait for some random time and closes request made by philosopher. Philosopher end up receiving an error and assumes fork not available and goes to **thinking** state
3. if philosopher manages to get right fork, then it tries for left fork (the `--fork2` one)
4. If left fork also available, philosopher starts **eating** process. If left fork is busy, after sometime it closes connection implying fork not available. Philosopher endup receiving an error assuming left fork not available. Since philosopher already got right fork before, philosopher returns right fork and then goes to **thinking** state
5. Once **thinking** done, philosopher retries to get forks starting whole process again

Fork request time can be configured by changing variable value `forkWaitDelay` in `fork.go`

Communication between fork and philosopher uses TCP. If communication fails, philosopher will panic

### Printing Process

Printing done by `stalker.go` process. All philosophers need to have passed stalker address when they started.
Must start `stalker.go` first before any philosopher. Process is as follows

1. When philosopher process starts, it sends **ping** message to stalker and stalker stores address of philosopher and seat number for later to decide messages are coming from which philosopher
2. Once dining starts, all philosophers sends a message **done** to stalker. stalker will print table header with seat numbers
3. For every message (**waiting**, **eating**, **thinking**, **done**) from philosopher, stalker will print message in respective column.

Communication between stalker and philosopher use UDP. any errors thrown by stalker will silently ignore by philosopher

### How to run locally ?

Each philosopher, fork and stalker (printer) needs to start separately. Can be in 1 machine or 10 machines
The number of cycles are hardcoded to 15 as variable `cycles`. Eating and thinking time is random which also hardcoded as variable `laziness`

1. Must to start stalker (printer) first to see register philosophers and its addresses for printing purpose

```
go run stalker.go --listen :5000
```

2. Then start 5 fork processes. port numbers can be any valid values

```
go run fork.go --listen :7001
go run fork.go --listen :7002
go run fork.go --listen :7003
go run fork.go --listen :7004
go run fork.go --listen :7005
```

3. Finally start philosophers. There should be one initiator (`--the-initiator`) to let others know that all philosophers joined at dining. Initiator will only tell when to eat.
   Please change **IP** and **Port** accordingly. Please make sure there can be only one `--the-initiator` else code will break

```shell
# start initiator with `--the-initiator`
go run philosopher.go --name Kant --listen :6001 --neighbour 127.0.0.1:6002 --fork 127.0.0.1:7001 --fork2 127.0.0.1:7002 --stalker 127.0.0.1:5000 --the-initiator

# remaining
go run philosopher.go --name Heidegger --listen :6002 --neighbour 127.0.0.1:6003 --fork 127.0.0.1:7002 --fork2 127.0.0.1:7003 --stalker 127.0.0.1:5000
go run philosopher.go --name Wittgenstein  --listen :6003 --neighbour 127.0.0.1:6004 --fork 127.0.0.1:7003 --fork2 127.0.0.1:7004 --stalker 127.0.0.1:5000
go run philosopher.go --name Locke  --listen :6004 --neighbour 127.0.0.1:6005 --fork 127.0.0.1:7004 --fork2 127.0.0.1:7005 --stalker 127.0.0.1:5000
go run philosopher.go --name Descartes  --listen :6005 --neighbour 127.0.0.1:6001 --fork 127.0.0.1:7005 --fork2 127.0.0.1:7001 --stalker 127.0.0.1:5000
```

### REFERENCES

- https://ipfs.io/ipfs/QmfYeDhGH9bZzihBUDEQbCbTc5k5FZKURMUoUvfmc27BwL/networkchannels/channels_of_channels.html
- https://www.linode.com/docs/development/go/developing-udp-and-tcp-clients-and-servers-in-go/
- https://github.com/doug/go-dining-philosophers
- https://en.wikipedia.org/wiki/Dining_philosophers_problem
- https://stackoverflow.com/questions/19992334/how-to-listen-to-n-channels-dynamic-select-statement
- https://stackoverflow.com/questions/61007385/golang-pattern-to-kill-multiple-goroutines-at-once
- https://stackoverflow.com/questions/52799280/golang-context-confusion-regarding-cancellation
- https://stackoverflow.com/questions/60122679/connections-stuck-at-close-wait-in-golang-server
