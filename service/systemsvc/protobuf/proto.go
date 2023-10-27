package protocol

//go:generate protoc --proto_path=$PWD:$PWD/.. --go_out=.  --go_opt=Mproto/system.proto=./pb/systempb     proto/system.proto
