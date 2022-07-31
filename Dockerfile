FROM golang:1.18-alpine as builder

WORKDIR /src

COPY . .

RUN apk add make openssl && make bin/procx_hostarch

FROM alpine:3.6 as runtime

COPY --from=builder /src/bin/procx_hostarch /bin/procx