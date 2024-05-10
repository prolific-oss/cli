FROM golang:1.22.3-alpine as builder
LABEL maintainer="Ben Selby <ben.selby@prolific.com>"

ENV APPNAME prolific
ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go

COPY . /go/src/github.com/prolific-oss/${APPNAME}

RUN apk update && \
	apk add --no-cache --virtual .build-deps \
	ca-certificates \
	gcc \
	libc-dev \
	libgcc \
	git \
	curl \
	make

RUN cd /go/src/github.com/prolific-oss/${APPNAME} && \
	make static-all  && \
	mv ${APPNAME} /usr/bin/${APPNAME}  && \
	apk del .build-deps  && \
	rm -rf /go

FROM scratch

COPY --from=builder /usr/bin/${APPNAME} /usr/bin/${APPNAME}
COPY --from=builder /etc/ssl/certs/ /etc/ssl/certs

ENV HOME /root

ENTRYPOINT [ "prolific" ]
CMD [ "--help"]
