FROM centurylink/ca-certs
MAINTAINER Dmitry Merkurev <didika914@gmail.com>
LABEL "com.centurylinklabs.watchtower"="true"

COPY watchtower /
ENTRYPOINT ["/watchtower"]
