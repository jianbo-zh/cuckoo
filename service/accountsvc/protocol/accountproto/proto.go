package accountproto

//go:generate protoc --proto_path=$PWD:$PWD/../../.. --go_out=. --go_opt=Mpb/account.proto=./pb pb/account.proto
