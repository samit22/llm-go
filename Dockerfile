FROM golang:1.23

WORKDIR /app

COPY . ./
RUN go mod download && \
CGO_ENABLED=0 GOOS=linux go build -o service .

EXPOSE 5000

ENV GIN_MODE=release

CMD ["./service"]
