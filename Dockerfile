FROM golang:1.23.4-alpine3.21 AS build
RUN apk add --no-cache git
# GOPRIVATE removed
WORKDIR /go/src/app/
COPY . .
RUN go build -o app-cmd
RUN go install github.com/go-swagger/go-swagger/cmd/swagger@v0.32.3
RUN swagger generate spec -o var/swagger.json --scan-models

FROM node:18-alpine AS openapi_convert
RUN npm install -g swagger2openapi
WORKDIR /convert
COPY --from=build /go/src/app/var/swagger.json /convert/swagger.json
RUN swagger2openapi -o /convert/openapi3.json /convert/swagger.json

FROM alpine:3.21
WORKDIR /opt/app
RUN apk add --no-cache tzdata
RUN mkdir /opt/app/var
ENV TZ=UTC
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

COPY --from=build /go/src/app/app-cmd /opt/app/
COPY --from=openapi_convert /convert/openapi3.json /opt/app/var/swagger.json
ADD migrations /opt/app/migrations
ADD var /opt/app/var
RUN chmod +x /opt/app/app-cmd
ENTRYPOINT ["/opt/app/app-cmd"]