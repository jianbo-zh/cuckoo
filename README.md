gomobile bind -o /Users/jianbo/workspace/my/study/android/Cuckoo/app/libs/chat.aar -target=android ./bind/core

go run github.com/tailscale/depaware --update ./bind/main


protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./bind/grpc/proto/chat.proto