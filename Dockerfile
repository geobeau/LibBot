FROM ubuntu:latest
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install -yq calibre
COPY libbot /bin/libbot


ENTRYPOINT [ "/bin/libbot" ]

