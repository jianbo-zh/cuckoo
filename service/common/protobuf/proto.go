package protobuf

//go:generate protoc --proto_path=$PWD:$PWD/.. --go_out=.  --go_opt=Mproto/session.proto=./pb/sessionpb proto/session.proto
