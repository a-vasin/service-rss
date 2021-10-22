#!/usr/bin/env bash

docker run                                             \
    --name "rss-db"                                       \
    --env POSTGRES_USER=postgres                       \
    --env POSTGRES_PASSWORD=postgres                   \
    --volume $(pwd)/db/sql:/docker-entrypoint-initdb.d \
    --detach                                           \
    --publish 5444:5432                                \
    postgres