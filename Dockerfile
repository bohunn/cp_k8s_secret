# Build Stage
FROM golang:1.19.1 AS builder

#WORKDIR /go/src/app
WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
#COPY . .
# Copy custom module
#COPY cluster /go/src/cluster

RUN go mod download

COPY main.go main.go
COPY cluster/ cluster/
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final Stage
#FROM alpine:latest
FROM gcr.io/distroless/static:nonroot


#RUN apk --no-cache add ca-certificates

WORKDIR /
#COPY --from=builder /go/src/app/main .
COPY --from=builder /workspace/main .

CMD ["/main", "-f", "/config.cfg"]
