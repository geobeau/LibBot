# FROM arm32v7/ubuntu:latest # If you want to build for ARM (rpi)
FROM amd64/ubuntu:latest
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install -yq calibre
COPY libbot /bin/libbot


ENTRYPOINT [ "/bin/libbot" ]

