FROM scratch

ADD /bin /bin

ENTRYPOINT ["/bin/aws-bootstrap"]
