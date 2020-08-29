protoc compact_formats.proto --go_out=plugins=grpc:.
protoc service.proto --go_out=plugins=grpc:.
protoc darkside.proto --go_out=plugins=grpc:.

