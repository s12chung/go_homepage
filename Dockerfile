FROM golang:1.10.3-alpine3.8

RUN echo '@testing http://nl.alpinelinux.org/alpine/edge/testing' >> /etc/apk/repositories &\
 apk --no-cache add\
 git make dep\
 nodejs nodejs-npm yarn\
 watchman@testing
RUN mkdir -p /var/run/watchman/root-state

RUN yarn global add webpack webpack-cli

WORKDIR /go/src/github.com/s12chung/go_homepage
COPY . .

RUN make install

EXPOSE 3000
EXPOSE 8080