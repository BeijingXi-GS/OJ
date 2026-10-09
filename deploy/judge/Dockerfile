FROM oj/compile-worker:gcc14.2.0-mvp AS tools
USER 0
RUN mkdir -p /source /solution /work /tmp && touch /source/solution.cpp /solution/program /input.txt && chmod 1777 /work /tmp
# Copy filesystem only: Redis' inherited VOLUME /data is deliberately absent.
FROM scratch
COPY --from=tools / /
COPY --chmod=0555 judge-compile-helper /usr/local/bin/judge-compile-helper
USER 10002:10002
ENV PATH=/usr/local/bin:/usr/bin:/bin
ENTRYPOINT ["/usr/local/bin/judge-compile-helper"]
