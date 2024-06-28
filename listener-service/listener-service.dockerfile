FROM golang:1.22.3-alpine as builder
RUN mkdir /app
COPY . /app
COPY go.mod /app
COPY go.sum /app
WORKDIR /app
RUN CGO_ENABLED=0 go build  -o listenerApp .
RUN chmod +x /app/listenerApp

FROM alpine:latest
RUN mkdir /app
COPY --from=builder /app/listenerApp /app
CMD ["/app/listenerApp"]