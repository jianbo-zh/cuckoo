package message

//go:generate protoc --proto_path=$PWD:$PWD/../../.. --go_out=. --go_opt=Mpb/peer_msg.proto=./pb pb/peer_msg.proto
//go:generate protoc --proto_path=$PWD:$PWD/../../.. --go_out=. --go_opt=Mpb/sync_peer_msg.proto=./pb pb/sync_peer_msg.proto
