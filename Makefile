gen:
	@protoc \
		--proto_path=protobuf "protobuf/mission_service.proto" \
		--go_out=pkg/genproto/missionservice --go_opt=paths=source_relative \
  		--go-grpc_out=pkg/genproto/missionservice --go-grpc_opt=paths=source_relative

	@protoc \
		--proto_path=protobuf "protobuf/auth_service.proto" \
		--go_out=pkg/genproto/authservice --go_opt=paths=source_relative \
  		--go-grpc_out=pkg/genproto/authservice --go-grpc_opt=paths=source_relative