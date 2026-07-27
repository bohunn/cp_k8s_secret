# Build Stage
FROM golang:1.26.5 AS builder

WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

COPY main.go main.go
COPY cluster/ cluster/
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final Stage
#FROM alpine:latest
FROM gcr.io/distroless/static:nonroot@sha256:f7f8f729987ad0fdf6b05eeeae94b26e6a0f613bdf46feea7fc40f7bd72953e6


#RUN apk --no-cache add ca-certificates

WORKDIR /
#COPY --from=builder /go/src/app/main .
COPY --from=builder /workspace/main .

CMD ["/main", "-f", "/config.cfg"]
