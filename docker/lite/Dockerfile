FROM debian:bookworm-slim as builder
RUN mkdir /src
COPY ./ /src
RUN /src/cp.sh

FROM scratch
COPY --from=builder /app/TorrServer /
WORKDIR /
ENTRYPOINT [ "/TorrServer" ]
