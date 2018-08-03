FROM golang:1.10.3-alpine3.8

ARG DOCKER_WORKDIR

# apk purposes for each line:
# go, webpack, aws, watchman (nice to have),
# imagemin-webpack-plugin, lovell/sharp image resizing (runtime), lovell/sharp image resizing (install/when yarn runs)
RUN echo '@testing http://nl.alpinelinux.org/alpine/edge/testing' >> /etc/apk/repositories &\
    apk --no-cache add\
    git make dep\
    nodejs nodejs-npm yarn\
    aws-cli@testing\
    watchman@testing\
    optipng gifsicle\
    fftw-dev@testing vips-dev@testing\
    g++ python2

RUN yarn global add webpack webpack-cli

RUN mkdir -p /var/run/watchman/root-state

EXPOSE 3000
EXPOSE 8080

WORKDIR $DOCKER_WORKDIR
COPY . .

RUN make install