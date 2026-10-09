FROM redis:7.4.5-alpine@sha256:bb186d083732f669da90be8b0f975a37812b15e913465bb14d845db72a4e3e08
RUN apk add --no-cache g++=14.2.0-r4 && addgroup -g 10002 compiler && adduser -D -u 10002 -G compiler compiler
COPY --chmod=0555 compile-worker /usr/local/bin/compile-worker
USER 10002:10002
ENV COMPILER_ADDR=:8081
EXPOSE 8081
ENTRYPOINT ["/usr/local/bin/compile-worker"]
