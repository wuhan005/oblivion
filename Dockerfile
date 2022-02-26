FROM alpine:latest
WORKDIR /home
COPY oblivion-server /home/oblivion-server
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
RUN chmod 655 oblivion-server
CMD ["/home/oblivion-server"]
EXPOSE 4000
