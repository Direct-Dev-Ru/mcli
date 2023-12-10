FROM golang:1.21-alpine as builder

WORKDIR /app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o mcli

# final stage

FROM alpine:latest

RUN apk add --update \
    curl \
    && rm -rf /var/cache/apk/*
# FROM scratch
# COPY --from=busybox:1.35 /bin/busybox /bin/busybox
# FROM busybox:stable-glibc

RUN mkdir -p /app

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