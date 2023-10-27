package protobuf

//go:generate protoc --proto_path=$PWD:$PWD/.. --go_out=. --go_opt=Mproto/account.proto=./pb/accountpb proto/account.proto
