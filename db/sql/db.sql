create table if not exists rss
(
    id                 serial primary key,
    email              text   not null,
    name               text   not null,
    sources            text[] not null,
    added_time         timestamp default now(),

    cached_rss         text,
    cached_valid_until timestamp,

    is_locked          bool      default false,
    locked_by          text,
    locked_time        timestamp
);

create index if not exists cached_valid_until_idx ON rss (cached_valid_until);

create unique index if not exists email_name_idx ON rss (email, name);