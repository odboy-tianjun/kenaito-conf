FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    bash curl jq ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /kenaito-conf-bash-client-demo

# Copy client library
COPY kenaito-conf-bash-client/ /kenaito-conf-bash-client/

# Copy demo
COPY kenaito-conf-bash-client-demo/app.sh /kenaito-conf-bash-client-demo/app.sh

RUN chmod +x app.sh

ENTRYPOINT ["bash", "app.sh"]
