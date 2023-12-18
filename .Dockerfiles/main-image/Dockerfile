FROM golang:1.21.5-bullseye as builder

WORKDIR /app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./

RUN go mod download && go mod verify

COPY . .

RUN go build -buildmode=plugin \
    -o plugins/http_default_plugins/http_plugins_compiled/http_default_handlers.so \
    plugins/http_default_plugins/src/default_http_handlers/http_default_handlers.go

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/mcli

# FROM busybox:stable-glibc as busybox
# final stage

FROM alpine:3.19.0
# FROM scratch
# FROM gcr.io/distroless/static-debian10

# COPY --from=busybox:1.35.0-uclibc /bin/sh /bin/sh
COPY --from=busybox:1.35.0-uclibc /bin/busybox /bin/busybox


# RUN apk add --no-cache curl && mkdir -p /app
    # && rm -rf /var/cache/apk/*

RUN busybox mkdir -p /app
    # && rm -rf /var/cache/apk/*

WORKDIR /app/

COPY ./http-data/ /app/http-data/

COPY ./http-static/ /app/http-static/

COPY --from=builder /app/mcli mcli 

COPY --from=builder /app/.mcli.yaml .mcli.yaml

COPY --from=builder /app/cors.json cors.json

ENV DEBUG=true

EXPOSE 8088

ENTRYPOINT ["./mcli"]

CMD ["-p", "8088"]