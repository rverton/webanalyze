FROM golang:1.25-alpine AS build-env
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
COPY cmd/webanalyze ./cmd/webanalyze
RUN CGO_ENABLED=0 go build -trimpath -o /webanalyze ./cmd/webanalyze

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app
RUN adduser -D webanalyze \
    && chown -R webanalyze:webanalyze /app
COPY --from=build-env /webanalyze .
USER webanalyze
RUN ["./webanalyze", "-update"]
ENTRYPOINT ["./webanalyze"]
