FROM golang:1.14.6-alpine3.12 AS builder

ARG BRANCH
ENV BRANCH $BRANCH

RUN mkdir /app 
ADD . /app/ 
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux go build -mod vendor -a -installsuffix cgo -o server .

FROM alpine:3.11
WORKDIR /

COPY --from=builder /app/server /server
COPY --from=builder /app/fastcheck.wsdl /fastcheck.wsdl
COPY --from=builder /app/fastcheck.xsd /fastcheck.xsd
ENTRYPOINT ["/server"]
