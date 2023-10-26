package protobuf

//go:generate protoc	--proto_path=$PWD:$PWD/..	--go_out=. 	--go_opt=Mproto/account.proto=./pb/accountpb 			proto/account.proto
//go:generate protoc	--proto_path=$PWD:$PWD/..	--go_out=. 	--go_opt=Mproto/contact.proto=./pb/contactpb 			proto/contact.proto
//go:generate protoc	--proto_path=$PWD:$PWD/..	--go_out=. 	--go_opt=Mproto/contact_msg.proto=./pb/contactpb 		proto/contact_msg.proto
//go:generate protoc	--proto_path=$PWD:$PWD/..	--go_out=. 	--go_opt=Mproto/contact_sync_msg.proto=./pb/contactpb 	proto/contact_sync_msg.proto
//go:generate protoc	--proto_path=$PWD:$PWD/..	--go_out=. 	--go_opt=Mproto/file.proto=./pb/filepb 					proto/file.proto
//go:generate protoc	--proto_path=$PWD:$PWD/..	--go_out=. 	--go_opt=Mproto/group_admin.proto=./pb/grouppb 			proto/group_admin.proto
//go:generate protoc	--proto_path=$PWD:$PWD/..	--go_out=. 	--go_opt=Mproto/group_admin_sync.proto=./pb/grouppb 	proto/group_admin_sync.proto
//go:generate protoc	--proto_path=$PWD:$PWD/..	--go_out=. 	--go_opt=Mproto/group_admin_pull.proto=./pb/grouppb 	proto/group_admin_pull.proto
//go:generate protoc	--proto_path=$PWD:$PWD/..	--go_out=. 	--go_opt=Mproto/group_msg_sync.proto=./pb/grouppb 		proto/group_msg_sync.proto
//go:generate protoc	--proto_path=$PWD:$PWD/..	--go_out=. 	--go_opt=Mproto/group_msg.proto=./pb/grouppb 			proto/group_msg.proto
//go:generate protoc	--proto_path=$PWD:$PWD/..	--go_out=. 	--go_opt=Mproto/group_connect.proto=./pb/grouppb 		proto/group_connect.proto
//go:generate protoc	--proto_path=$PWD:$PWD/..	--go_out=. 	--go_opt=Mproto/group_routing_table.proto=./pb/grouppb 	proto/group_routing_table.proto
//go:generate protoc	--proto_path=$PWD:$PWD/..	--go_out=. 	--go_opt=Mproto/session.proto=./pb/sessionpb 			proto/session.proto
//go:generate protoc	--proto_path=$PWD:$PWD/..	--go_out=. 	--go_opt=Mproto/deposit.proto=./pb/depositpb 			proto/deposit.proto
//go:generate protoc	--proto_path=$PWD:$PWD/..	--go_out=. 	--go_opt=Mproto/system.proto=./pb/systempb 				proto/system.proto
