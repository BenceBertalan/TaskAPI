FROM golang:latest
RUN go get github.com/gorilla/mux && go get gopkg.in/mgo.v2
COPY main.go /app/
COPY config.json /app/
WORKDIR /app/
CMD ["go", "run", "/app/main.go"]