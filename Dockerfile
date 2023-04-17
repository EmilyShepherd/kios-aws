FROM alpine

ADD /bin /bin
ADD /assets /etc/templates

ENTRYPOINT ["/bin/aws-bootstrap"]
