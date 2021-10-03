create table rss
(
    id                 serial primary key,
    name               text unique not null,
    sources            text[]      not null,

    cached_rss         text,
    cached_valid_until timestamp,

    is_locked          bool default false,
    locked_by          text,
    locked_time        timestamp
);

create index cached_valid_until_idx ON rss (cached_valid_until);