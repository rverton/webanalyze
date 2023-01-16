FROM golang:latest AS build-env
RUN go install github.com/rverton/webanalyze/cmd/webanalyze@latest
FROM alpine:latest
RUN apk add --no-cache libc6-compat
WORKDIR /app
COPY --from=build-env /go/bin/webanalyze .
RUN mkdir -p /app \
    && adduser -D webanalyze \
    && chown -R webanalyze:webanalyze /app
USER webanalyze
RUN ["./webanalyze", "-update"]
ENTRYPOINT ["./webanalyze"]