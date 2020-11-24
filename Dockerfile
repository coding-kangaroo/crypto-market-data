FROM golang
COPY . /tmp
WORKDIR /tmp
ENTRYPOINT ["make", "all"]