FROM golang:1.11-stretch

RUN mkdir /app 
ADD . /app/ 
WORKDIR /app 


RUN go get -u  "github.com/mongodb/mongo-go-driver/bson"
##RUN go get -u	"github.com/mongodb/mongo-go-driver/bson/objectid"
RUN go get -u	"github.com/mongodb/mongo-go-driver/mongo"

RUN CGO_ENABLED=1 GOOS=js GOARCH=wasm go get -u "github.com/dennwc/dom"
RUN CGO_ENABLED=1 GOOS=js GOARCH=wasm go get -u "syscall/js" 

RUN CGO_ENABLED=1 GOARCH=wasm GOOS=js go build -o test.wasm main.go

RUN CGO_ENABLED=1 GOARCH=amd64 GOOS=linux go build -o server server.go

EXPOSE 3000

CMD ["./server"]


