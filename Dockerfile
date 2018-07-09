FROM golang:1.10.3-alpine3.8

RUN apk --no-cache add\
 git make dep\
 nodejs nodejs-npm yarn
RUN yarn global add webpack webpack-cli

WORKDIR /go/src/github.com/s12chung/go_homepage
COPY . .

RUN make install

EXPOSE 3000