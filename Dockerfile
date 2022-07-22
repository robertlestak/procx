FROM golang:1.18 as builder

WORKDIR /src

COPY . .

RUN go build -o /bin/qjob cmd/qjob/*.go

FROM alpine:3.6 as runtime

COPY --from=builder /bin/qjob /bin/qjob