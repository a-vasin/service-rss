# syntax=docker/dockerfile:1

## Build
FROM golang:1.16 AS build

WORKDIR /app
COPY . /app

RUN go build -o /app/bin/service-entrypoint /app/cmd/service

## Deploy
FROM gcr.io/distroless/base-debian10

WORKDIR /

COPY --from=build /app/bin/service-entrypoint /service
COPY ./jsonschema/ /jsonschema/
COPY ./html/ /html/

EXPOSE 80

CMD [ "/service" ]