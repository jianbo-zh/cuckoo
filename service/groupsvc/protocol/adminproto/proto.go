package adminproto

//go:generate protoc --proto_path=$PWD:$PWD/../../.. --go_out=. --go_opt=Mpb/group_log.proto=./pb pb/group_log.proto
//go:generate protoc --proto_path=$PWD:$PWD/../../.. --go_out=. --go_opt=Mpb/group_log_sync.proto=./pb pb/group_log_sync.proto
