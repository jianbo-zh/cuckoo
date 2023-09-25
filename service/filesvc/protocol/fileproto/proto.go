package fileproto

//go:generate protoc --proto_path=$PWD:$PWD/../../.. --go_out=. --go_opt=Mpb/file.proto=./pb pb/file.proto
