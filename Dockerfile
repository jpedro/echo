FROM scratch

COPY echoes-linux-amd64 /srv/echoes

ENTRYPOINT ["/srv/echoes"]
