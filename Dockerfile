FROM scratch

WORKDIR /srv
COPY echo-linux-amd64 /srv/echo

EXPOSE 8080

CMD ["/srv/echo"]
