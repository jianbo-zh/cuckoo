package protocol

//go:generate protoc --proto_path=$PWD:$PWD/.. --go_out=.  --go_opt=Mproto/contact.proto=./pb/contactpb    proto/contact.proto
//go:generate protoc --proto_path=$PWD:$PWD/.. --go_out=.  --go_opt=Mproto/contact_msg.proto=./pb/contactpb   proto/contact_msg.proto
//go:generate protoc --proto_path=$PWD:$PWD/.. --go_out=.  --go_opt=Mproto/contact_sync_msg.proto=./pb/contactpb  proto/contact_sync_msg.proto
