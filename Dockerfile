FROM scratch

COPY echo-linux-amd64 /srv/echo

CMD ["/srv/echo"]
