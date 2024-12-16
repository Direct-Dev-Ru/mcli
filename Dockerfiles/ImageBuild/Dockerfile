FROM --platform=${BUILDPLATFORM} golang:1.23-alpine AS builder

ARG TARGETARCH

# RUN apk add git && go install mvdan.cc/garble@latest

WORKDIR /app

COPY . .

RUN echo $BUILDPLATFORM > buildplatform
RUN echo $TARGETARCH > targetarch

# RUN GOOS=linux GOARCH=$TARGETARCH go build -o /app/go-lcg .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-s -w" -o /app/swknf .
# RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} garble -literals -tiny build -ldflags="-w -s" -o /app/swknf .

FROM alpine:latest

WORKDIR /root

# COPY --from=builder /app/buildplatform .
# COPY --from=builder /app/targetarch .
COPY --from=builder /app/swknf /root/swknf

ENTRYPOINT ["/root/swknf"]