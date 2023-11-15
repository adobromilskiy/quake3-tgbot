FROM golang:1.21-alpine AS builder
ADD . /build
WORKDIR /build
RUN cd app && go build -v -mod=vendor -o /app
FROM alpine AS final
COPY --from=builder /app /app
ENTRYPOINT [ "/app" ]