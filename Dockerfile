FROM scratch

COPY --from=alpine:3.20 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY lootsheet-raspi /lootsheet

ENV XDG_DATA_HOME=/data
ENV XDG_CONFIG_HOME=/data

EXPOSE 7547

ENTRYPOINT ["/lootsheet", "serve", "--no-tls", "--addr", ":7547"]
