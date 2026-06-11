FROM golang:1.23.4-alpine3.21 as build
RUN apk add --no-cache git
WORKDIR /go/src/app/
COPY . .
RUN go build -o app-cmd

FROM quay.io/goswagger/swagger as swaggerGenerate
WORKDIR /go/src/app/
COPY . .
RUN GOPATH=/go swagger generate spec -o var/swagger.json

FROM alpine:3.21
RUN apk update
RUN apk upgrade
RUN apk add curl busybox-extras shadow supervisor
RUN mkdir /etc/supervisor.d/
COPY docker/supervisor/supervisor.conf /etc/supervisor.d/supervisor.ini
RUN useradd --create-home -s /bin/sh app
#RUN chown -R app:app /opt/app
#RUN chmod -R 770 /app
#USER app
WORKDIR /opt/app
RUN mkdir /opt/app/var
RUN chown -R app:app /opt/app
COPY --chown=app:app --from=build /go/src/app/app-cmd /opt/app/
COPY --chown=app:app --from=swaggerGenerate /go/src/app/var/swagger.json /opt/app/var/swagger.json
ADD --chown=app:app migrations /opt/app/migrations
ADD --chown=app:app html/static /opt/app/html/static
ADD --chown=app:app html/templates /opt/app/html/templates
ADD --chown=app:app var /opt/app/var
RUN chmod +x /opt/app/app-cmd

ENV TZ "UTC"

#ENTRYPOINT ["/opt/app/app-cmd"]
ENTRYPOINT ["/usr/bin/supervisord"]

HEALTHCHECK --start-period=10s --interval=10s --timeout=5s --retries=3 CMD curl -f http://127.0.0.1:8090/_/liveness || exit 1

EXPOSE 8090


