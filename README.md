# grpc-gateway
grpc-gateway learning

## 安装

```shell script
go get -u google.golang.org/grpc
go get -u github.com/golang/protobuf/protoc-gen-go
go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
```

## hello world

### hello.proto

```go
syntax = "proto3";

package proto;

import "google/api/annotations.proto";

service HelloWorld {
    rpc SayHelloWorld (HelloWorldRequest) returns (HelloWorldResponese) {
        option (google.api.http) = {
            post: "/hello_world"
            body: "*"
        };
    }
}

message HelloWorldRequest {
    string referer = 1;
}

message HelloWorldResponese {
    string message = 1;
}
```

### 生成代码

利用go generate或者直接执行命令：

```go
package proto

// generate pb
//go:generate protoc  -I/usr/local/include -I. -I$GOPATH/src -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --go_out=plugins=grpc:. hello.proto

// generate gateway reverse-proxy
//go:generate protoc -I/usr/local/include -I. -I$GOPATH/src -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --grpc-gateway_out=logtostderr=true:. hello.proto

// generate swagger doc
//go:generate protoc -I/usr/local/include -I. -I$GOPATH/src -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --swagger_out=logtostderr=true:. hello.proto

```

### grpc-gateway mux

```go
func newGateway() http.Handler {
	// grpc-gateway service
	ctx := context.Background()
	dcreds, err := credentials.NewClientTLSFromFile(CertPemPath, CertName)
	if err != nil {
		log.Printf("Failed to create client TLS credentials %v\n", err)
	}
	dopts := []grpc.DialOption{grpc.WithTransportCredentials(dcreds)}
	gwmux := runtime.NewServeMux()

	// register grpc-gateway pb
	if err := pb.RegisterHelloWorldHandlerFromEndpoint(ctx, gwmux, EndPoint, dopts); err != nil {
		log.Fatalf("Failed to register gw server: %v\n", err)
	}
	return gwmux
}
```



### grpc与http分流

```go
func GrpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	if otherHandler == nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			grpcServer.ServeHTTP(w, r)
		})
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	})
}
```

## 使用yaml配置

### proto

```go
syntax = "proto3";

package proto;


service HelloWorld {
    rpc SayHelloWorld (HelloWorldRequest) returns (HelloWorldResponese) {
    }
}

message HelloWorldRequest {
    string referer = 1;
}

message HelloWorldResponese {
    string message = 1;
}
```

### yaml

```yaml
type: google.api.Service
config_version: 3

http:
  rules:
    - selector: proto.HelloWorld.SayHelloWorld
      post: /hello_world
      body: "*"
```

### 生成代码

```go
// 以下为使用yaml配置，代码生成
// generate pb
//go:generate protoc  -I/usr/local/include -I. -I$GOPATH/src -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --go_out=plugins=grpc:. hello.proto

// generate gateway reverse-proxy
//go:generate protoc -I/usr/local/include -I. -I$GOPATH/src -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --grpc-gateway_out=logtostderr=true,grpc_api_configuration=./hello.yaml:. hello.proto

// generate swagger doc
//go:generate protoc -I/usr/local/include -I. -I$GOPATH/src -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --swagger_out=logtostderr=true,grpc_api_configuration=./hello.yaml:. hello.proto

```



