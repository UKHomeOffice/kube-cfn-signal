FROM alpine:3.4

RUN apk upgrade --no-cache
RUN apk add --no-cache ca-certificates

COPY bin/kube-cfn-signal_linux_amd64 /bin/kube-cfn-signal

ENTRYPOINT ["/bin/kube-cfn-signal"]
