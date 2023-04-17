FROM alpine

ADD /bin /bin

ENTRYPOINT ["/bin/aws-bootstrap"]
