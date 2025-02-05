FROM golang:1.22.3-alpine as builder
RUN mkdir /app
COPY . /app
COPY go.mod /app
COPY go.sum /app
WORKDIR /app
RUN go mod tidy
RUN CGO_ENABLED=0 go build  -o brokerApp ./cmd/api
RUN chmod +x /app/brokerApp

FROM alpine:latest
RUN mkdir /app
COPY --from=builder /app/brokerApp /app
CMD ["/app/brokerApp"]