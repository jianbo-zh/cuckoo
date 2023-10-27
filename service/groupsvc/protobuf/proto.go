package protocol

//go:generate protoc	--proto_path=$PWD:$PWD/..	--go_out=. 	--go_opt=Mproto/group_admin.proto=./pb/grouppb 			proto/group_admin.proto
//go:generate protoc	--proto_path=$PWD:$PWD/..	--go_out=. 	--go_opt=Mproto/group_admin_sync.proto=./pb/grouppb 	proto/group_admin_sync.proto
//go:generate protoc	--proto_path=$PWD:$PWD/..	--go_out=. 	--go_opt=Mproto/group_admin_pull.proto=./pb/grouppb 	proto/group_admin_pull.proto
//go:generate protoc	--proto_path=$PWD:$PWD/..	--go_out=. 	--go_opt=Mproto/group_msg_sync.proto=./pb/grouppb 		proto/group_msg_sync.proto
//go:generate protoc	--proto_path=$PWD:$PWD/..	--go_out=. 	--go_opt=Mproto/group_msg.proto=./pb/grouppb 			proto/group_msg.proto
//go:generate protoc	--proto_path=$PWD:$PWD/..	--go_out=. 	--go_opt=Mproto/group_connect.proto=./pb/grouppb 		proto/group_connect.proto
//go:generate protoc	--proto_path=$PWD:$PWD/..	--go_out=. 	--go_opt=Mproto/group_routing_table.proto=./pb/grouppb 	proto/group_routing_table.proto
