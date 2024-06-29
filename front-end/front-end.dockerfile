FROM golang:1.22.3-alpine as builder
RUN mkdir /app
COPY . /app
COPY go.mod /app
WORKDIR /app
RUN go mod tidy
RUN CGO_ENABLED=0 go build  -o frontEndApp ./cmd/web
RUN chmod +x /app/frontEndApp

FROM alpine:latest
RUN mkdir /app
COPY --from=builder /app/frontEndApp /app
CMD ["/app/frontEndApp"]