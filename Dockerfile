FROM golang:1.17.5-alpine as builder

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /go/src/setup-krew

COPY go.mod go.sum ./

RUN  go mod download

COPY ./ ./

RUN go build -ldflags='-s -w'

FROM gcr.io/distroless/static:debug

COPY --from=builder /go/src/setup-krew/setup-krew /usr/local/bin/setup-krew

ENTRYPOINT [ "setup-krew" ]
