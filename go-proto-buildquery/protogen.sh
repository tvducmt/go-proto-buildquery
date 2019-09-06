protoc -I/usr/local/include -I. \
  -I$GOPATH/src \
  -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
  -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway \
  -I${GOPATH}/src/github.com/tvducmt/go-proto-buildquery \
  --go_out=plugins=grpc:. --grpc-gateway_out=logtostderr=true:. \
  --buildquery_out="lang=go:." \
  *.proto 