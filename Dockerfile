FROM alpine:latest

COPY libbot /bin/libbot

ENTRYPOINT [ "/bin/libbot" ]

