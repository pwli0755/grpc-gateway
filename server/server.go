package server

import (
	"context"
	"crypto/tls"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"grpc_gateway/pkg/util"
	pb "grpc_gateway/proto"
	"log"
	"net"
	"net/http"
	"path"
	"strings"
)

var (
	ServerPort  string
	CertName    string
	CertPemPath string
	CertKeyPath string
	EndPoint    string
	SwaggerDir  string

	tlsConfig *tls.Config
)

func Serve() (err error) {
	EndPoint = ":" + ServerPort
	conn, err := net.Listen("tcp", EndPoint)
	if err != nil {
		log.Fatalf("TCP listen err: %v", err)
	}
	tlsConfig, err := util.GetTLSConfig(CertPemPath, CertKeyPath)
	if err != nil {
		log.Fatalf("GetTLSConfig err: %v", err)
	}
	srv := newServer(tlsConfig)
	log.Printf("gRPC and https listen on: %s\n", ServerPort)
	if err = srv.Serve(tls.NewListener(conn, tlsConfig)); err != nil {
		log.Fatalf("ListenAndServe: %v\n", err)
	}
	return err
}

func newServer(tlsConfig *tls.Config) *http.Server {
	grpcServer := newGrpc()

	gwmux := newGateway()

	// http服务
	mux := http.NewServeMux()
	mux.Handle("/", gwmux)
	mux.HandleFunc("/swagger/", serveSwaggerFile)
	serveSwaggerUI(mux)

	return &http.Server{
		Addr:      EndPoint,
		Handler:   util.GrpcHandlerFunc(grpcServer, mux),
		TLSConfig: tlsConfig,
	}
}

func serveSwaggerUI(mux *http.ServeMux) {
	//fileServer := http.FileServer(&assetfs.AssetFS{
	//	Asset: swagger.Asset,
	//	AssetDir: swagger.AssetDir,
	//	Prefix: "third_party/swagger-ui",
	//})
	prefix := "/swaggerUI/"
	mux.Handle(prefix, http.StripPrefix(prefix, http.FileServer(http.Dir("third_party/swagger-ui/"))))
}

func serveSwaggerFile(w http.ResponseWriter, r *http.Request) {
	if !strings.HasSuffix(r.URL.Path, "swagger.json") {
		log.Printf("Not Found: %s", r.URL.Path)
		http.NotFound(w, r)
		return
	}
	p := strings.TrimPrefix(r.URL.Path, "/swagger/")
	p = path.Join(SwaggerDir, p)

	log.Printf("Serving swagger-file: %s\n", p)

	http.ServeFile(w, r, p)
}

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

func newGrpc() *grpc.Server {
	var opts []grpc.ServerOption

	// grpc server
	creds, err := credentials.NewClientTLSFromFile(CertPemPath, CertKeyPath)
	if err != nil {
		log.Fatalf("Failed to create server TLS credentials %v", err)
	}
	opts = append(opts, grpc.Creds(creds))
	grpcServer := grpc.NewServer(opts...)

	// register grpc pb
	pb.RegisterHelloWorldServer(grpcServer, NewHelloService())
	return grpcServer
}
