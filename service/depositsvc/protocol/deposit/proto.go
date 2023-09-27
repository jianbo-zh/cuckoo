package deposit

//go:generate protoc --proto_path=$PWD:$PWD/../../.. --go_out=. --go_opt=Mpb/offline_msg.proto=./pb pb/offline_msg.proto
