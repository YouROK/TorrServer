FROM alpine
LABEL maintainer "9000000"
RUN apk add --no-cache wget
COPY start.sh /start.sh
RUN chmod +x /start.sh
ENTRYPOINT ["/bin/sh", "/start.sh"]
