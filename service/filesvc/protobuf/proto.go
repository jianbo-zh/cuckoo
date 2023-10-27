package protocol

//go:generate protoc  --proto_path=$PWD:$PWD/..  --go_out=.   --go_opt=Mproto/file.proto=./pb/filepb           proto/file.proto
