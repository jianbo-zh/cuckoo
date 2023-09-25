package contactproto

//go:generate protoc --proto_path=$PWD:$PWD/../../.. --go_out=. --go_opt=Mpb/contact.proto=./pb pb/contact.proto
