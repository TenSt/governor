FROM golang:1.12.13-stretch

RUN mkdir /app 
ADD . /app/ 
WORKDIR /app 

RUN go get -u "go.mongodb.org/mongo-driver/bson"
RUN go get -u "go.mongodb.org/mongo-driver/mongo"

RUN CGO_ENABLED=1 GOARCH=amd64 GOOS=linux go build -o server server.go

EXPOSE 3000

CMD ["./server"]