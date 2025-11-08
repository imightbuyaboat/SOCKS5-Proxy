create table if not exists users (
    id serial primary key,
    username text unique not null,
    password_hash text not null
);