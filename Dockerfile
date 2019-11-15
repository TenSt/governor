FROM golang:1.12.13-stretch

RUN mkdir /app 
ADD . /app/ 
WORKDIR /app 

RUN go get -u "github.com/auth0/go-jwt-middleware"
RUN go get -u "github.com/codegangsta/negroni"
RUN go get -u "github.com/dgrijalva/jwt-go"

RUN go get -u "go.mongodb.org/mongo-driver/bson"
RUN go get -u "go.mongodb.org/mongo-driver/mongo"
RUN go get -u "github.com/buger/jsonparser"

#RUN go get -u "syscall/js" 
RUN go get -u "github.com/dennwc/dom"

RUN CGO_ENABLED=1 GOARCH=wasm GOOS=js go build -o test.wasm main.go

RUN CGO_ENABLED=1 GOARCH=amd64 GOOS=linux go build -o server server.go

EXPOSE 3000

CMD ["./server"]


