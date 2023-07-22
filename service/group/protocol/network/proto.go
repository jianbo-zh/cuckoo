package network

//go:generate protoc --proto_path=$PWD:$PWD/../../.. --go_out=. --go_opt=Mpb/routing_table.proto=./pb pb/routing_table.proto
//go:generate protoc --proto_path=$PWD:$PWD/../../.. --go_out=. --go_opt=Mpb/connect.proto=./pb pb/connect.proto
