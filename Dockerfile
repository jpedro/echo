FROM scratch

COPY echo-linux-amd64 /srv/echo
COPY version.txt /srv/

ENTRYPOINT ["/srv/echo"]
