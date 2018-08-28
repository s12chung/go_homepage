FROM golang:1.10.3-alpine3.8

ARG DOCKER_WORKDIR

# apk purposes for each line:
# go, webpack, aws
# imagemin-webpack-plugin, lovell/sharp image resizing (runtime), lovell/sharp image resizing (install/when yarn runs)
RUN echo '@testing http://nl.alpinelinux.org/alpine/edge/testing' >> /etc/apk/repositories &\
    apk --no-cache add\
    git make dep\
    nodejs nodejs-npm yarn\
    aws-cli@testing\
    optipng gifsicle\
    fftw-dev@testing vips-dev@testing\
    g++ python2

RUN yarn global add webpack webpack-cli

RUN mkdir -p /var/run/watchman/root-state

EXPOSE 3000
EXPOSE 8080

WORKDIR $DOCKER_WORKDIR
COPY . .

# install watchman from custom build because of: https://github.com/facebook/watchman/issues/602
RUN apk add ./watchman/watchman-4.7.0-r0.apk --allow-untrusted

RUN make install