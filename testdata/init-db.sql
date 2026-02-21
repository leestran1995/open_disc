create schema open_discord;

create sequence open_discord.rooms_default_order;

alter sequence open_discord.rooms_default_order owner to appuser;

create table if not exists open_discord.rooms
(
    id         uuid    default gen_random_uuid() not null
        constraint rooms_pk
            primary key,
    name       varchar(255)                      not null
        constraint server_name_unique
            unique,
    sort_order integer default nextval('open_discord.rooms_default_order'::regclass)
);

alter table open_discord.rooms
    owner to appuser;

alter sequence open_discord.rooms_default_order owned by open_discord.rooms.sort_order;

create table if not exists open_discord.users
(
    id       uuid default gen_random_uuid() not null
        constraint users_pk
            primary key,
    nickname varchar(256)
        constraint users_nickname_unique
            unique,
    username varchar(256)
        constraint users_pk_2
            unique,
    password text
);

alter table open_discord.users
    owner to appuser;

create table if not exists open_discord.messages
(
    id        uuid                     default gen_random_uuid() not null
        constraint messages_pk
            primary key,
    timestamp timestamp with time zone default CURRENT_TIMESTAMP not null,
    room_id   uuid                                               not null
        constraint messages_server_id_fk
            references open_discord.rooms,
    message   text,
    username  varchar(256)                                       not null
        constraint messages_users_username_fk
            references open_discord.users (username)
);

alter table open_discord.messages
    owner to appuser;

create index if not exists messages_timestamp_index
    on open_discord.messages (timestamp desc);

