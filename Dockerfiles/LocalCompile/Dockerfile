FROM --platform=${BUILDPLATFORM} golang:1.23-alpine AS build
ARG TARGETOS
ARG TARGETARCH
# RUN apk add git && go install mvdan.cc/garble@latest
WORKDIR /src
ENV CGO_ENABLED=0
# COPY go.* .
# RUN go mod download
COPY . .

# RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /out/swknf .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-s -w" -o /out/swknf .
# RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} garble  -literals -tiny build -ldflags="-w -s" -o /out/swknf .

FROM scratch AS bin-unix
COPY --from=build /out/swknf /swknf

FROM bin-unix AS bin-linux
FROM bin-unix AS bin-darwin

FROM scratch AS bin-windows
COPY --from=build /out/swknf /swknf.exe

FROM bin-${TARGETOS} AS bin