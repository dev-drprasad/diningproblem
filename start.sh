go run philosopher.go --name Kant --listen :6001 --fork 127.0.0.1:7001 --neighbour 127.0.0.1:6002 --the-initiator --fork2 127.0.0.1:7002 --stalker 127.0.0.1:5000
go run philosopher.go --name Heidegger --listen :6002 --fork 127.0.0.1:7002 --neighbour 127.0.0.1:6003 --fork2 127.0.0.1:7003 --stalker 127.0.0.1:5000
go run philosopher.go --name Wittgenstein  --listen :6003 --fork 127.0.0.1:7003 --neighbour 127.0.0.1:6004 --fork2 127.0.0.1:7004 --stalker 127.0.0.1:5000
go run philosopher.go --name Locke  --listen :6004 --fork 127.0.0.1:7004 --neighbour 127.0.0.1:6005 --fork2 127.0.0.1:7005 --stalker 127.0.0.1:5000
go run philosopher.go --name Descartes  --listen :6005 --fork 127.0.0.1:7005 --neighbour 127.0.0.1:6001 --fork2 127.0.0.1:7001 --stalker 127.0.0.1:5000



go run fork.go --listen :7001
go run fork.go --listen :7002
go run fork.go --listen :7003
go run fork.go --listen :7004
go run fork.go --listen :7005
