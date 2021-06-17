FROM scratch

COPY echo /srv/echo
COPY version.txt /srv/

ENTRYPOINT ["/srv/echo"]
