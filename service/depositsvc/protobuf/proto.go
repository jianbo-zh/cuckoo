package protocol

//go:generate protoc --proto_path=$PWD:$PWD/.. --go_out=.  --go_opt=Mproto/deposit.proto=./pb/depositpb proto/deposit.proto
