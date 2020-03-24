FROM pangpanglabs/golang:builder AS builder
WORKDIR /go/src/github.com/hublabs/order-api
COPY . .
# disable cgo
ENV CGO_ENABLED=0
# build steps
RUN echo ">>> 1: go version" && go version
RUN echo ">>> 2: go get" && go get -v -d
RUN echo ">>> 3: go install" && go install

# make application docker image use alpine
FROM pangpanglabs/alpine-ssl
# using timezone
WORKDIR /go/bin/
# copy config files to image
COPY --from=builder /go/src/github.com/hublabs/order-api/*.yml ./
# COPY --from=builder /swagger-ui/ ./swagger-ui/
# copy execute file to image
COPY --from=builder /go/bin/order-api ./
EXPOSE 8000
CMD ["./order-api"]
