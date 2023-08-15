FROM --platform=${BUILDPLATFORM} golang:1.21-alpine AS build
ARG TARGETOS
ARG TARGETARCH
WORKDIR /src
ENV CGO_ENABLED=0
COPY go.* .
RUN go mod download
COPY . .
# RUN echo ${TARGETOS} && echo ${TARGETARCH}
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /out/mcli .

FROM scratch AS bin-unix
COPY --from=build /out/mcli /

FROM bin-unix AS bin-linux
FROM bin-unix AS bin-darwin

FROM scratch AS bin-windows
COPY --from=build /out/mcli /mcli.exe

FROM bin-${TARGETOS} AS bin