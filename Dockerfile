FROM scratch

COPY lootsheet-raspi /lootsheet

ENV XDG_DATA_HOME=/data
ENV XDG_CONFIG_HOME=/data

VOLUME /data/lootsheet

EXPOSE 7547

ENTRYPOINT ["/lootsheet", "serve", "--no-tls", "--addr", ":7547"]
