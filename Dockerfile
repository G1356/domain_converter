FROM docker.io/traefik:v3.5.4
COPY . /plugins-local/src/github.com/G1356/domain_converter
COPY plugin-simplecache /plugins-local/src/github.com/traefik/plugin-simplecache