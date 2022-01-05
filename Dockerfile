FROM golang:1.16-alpine3.13 AS builder
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories
RUN apk add --no-cache --virtual build-deps make git openssh g++

# setup go env
ENV GO111MODULE=on
ENV GOPROXY="https://goproxy.cn,direct"
ENV GOSUMDB="sum.golang.google.cn"

WORKDIR /build
COPY . .
RUN go build -tags musl -o kafka-benchmark ./...

FROM alpine:3.13
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories
RUN apk --no-cache add tzdata
ENV TZ=Asia/Shanghai

# set up nsswitch.conf for Go's "netgo" implementation
# https://github.com/golang/go/issues/35305
RUN `[ ! -e /etc/nsswitch.conf ] && echo 'hosts: files dns' > /etc/nsswitch.conf`

WORKDIR /app
COPY --from=builder /build/kafka-benchmark .

# setup your app to run
ENTRYPOINT ["./kafka-benchmark"]
CMD []
