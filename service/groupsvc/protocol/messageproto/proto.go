package messageproto

//go:generate protoc --proto_path=$PWD:$PWD/../../.. --go_out=. --go_opt=Mpb/group_msg.proto=./pb pb/group_msg.proto
//go:generate protoc --proto_path=$PWD:$PWD/../../.. --go_out=. --go_opt=Mpb/group_msg_sync.proto=./pb pb/group_msg_sync.proto
