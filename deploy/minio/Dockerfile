# The cached, pinned official image supplies CA certificates only.
ARG RUNTIME_BASE=redis:7.4.5-alpine@sha256:bb186d083732f669da90be8b0f975a37812b15e913465bb14d845db72a4e3e08
FROM ${RUNTIME_BASE} AS certificates

FROM scratch
COPY --from=certificates /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY minio /minio
COPY healthcheck /healthcheck
COPY LICENSE /licenses/minio-LICENSE
COPY --chown=10001:10001 data/ /data/
USER 10001:10001
WORKDIR /data
EXPOSE 9000 9001
ENTRYPOINT ["/minio"]
CMD ["server", "/data", "--console-address", ":9001"]
