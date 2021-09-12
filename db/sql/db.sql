create table rss
(
    id      serial primary key,
    name    text unique not null,
    sources text[]      not null
)