

android:
	@gomobile bind -o /Users/jianbo/workspace/my/study/android/Cuckoo/app/libs/chat.aar -target=android ./bind/core

.PHONY: android 


proto:
	@protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./bind/grpc/proto/chat.proto

.PHONY: proto 