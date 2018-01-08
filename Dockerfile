FROM golang:alpine as builder
LABEL MAINTAINER="Michael Gasch <embano1@live.com>"
ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go

COPY . /go/src/github.com/embano1/tw

RUN	apk add --no-cache \
	ca-certificates

RUN set -x \
	&& apk add --no-cache --virtual .build-deps \
		git \
		gcc \
		libc-dev \
		libgcc \
		make \
	&& cd /go/src/github.com/embano1/tw \
	&& make static \
	&& mv tw /usr/bin/tw \
	&& apk del .build-deps \
	&& rm -rf /go \
	&& echo "Build complete."

FROM scratch

COPY --from=builder /usr/bin/tw /usr/bin/tw
COPY --from=builder /etc/ssl/certs/ /etc/ssl/certs

ENTRYPOINT [ "tw" ]
CMD [ "--help" ]