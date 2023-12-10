FROM golang:1.21-alpine as builder

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o mcli

# final stage

#FROM alpine:latest
FROM scratch
COPY --from=busybox:1.35 /bin/busybox /bin/busybox
WORKDIR /app/
# COPY --from=builder /usr/src/app .
COPY --from=builder /usr/src/app/mcli mcli 
COPY --from=builder /usr/src/app/.mcli.yaml .mcli.yaml
ENTRYPOINT [ "./mcli"]

