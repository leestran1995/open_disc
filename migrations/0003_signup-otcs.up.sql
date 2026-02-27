create table open_discord.signup_otcs (
    code uuid not null,
    used bool default false,
    time_created timestamp default current_timestamp,
    time_expires timestamp default current_timestamp + interval '1 hour'
);