FROM scratch

COPY echo-linux-amd64 /srv/echo

ENTRYPOINT ["/srv/echo"]
CMD ["serve", "--env", "prod"]
