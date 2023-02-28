FROM golang:1.20.0-alpine as builder

ENV GO111MODULE=on

WORKDIR /build

COPY src/ .

RUN go build -ldflags="-s -w" -o assets/out out/out.go out/structs.go && \
    go build -ldflags="-s -w" -o assets/in in/in.go in/structs.go

RUN apk add --update --no-cache ca-certificates

FROM scratch

COPY --from=builder --chmod=744 /build/assets/out /opt/resource/out
COPY --from=builder --chmod=744 /build/assets/in /opt/resource/in
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --chmod=744 src/check/check.sh /opt/resource/check