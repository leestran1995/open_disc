create table open_discord.roles (
    id uuid not null default gen_random_uuid(),
    name varchar(128) not null,
    primary key (id)
);

create table open_discord.user_roles (
    user_id uuid not null references open_discord.users (id),
    role_id uuid not null references open_discord.roles (id),
    primary key (user_id, role_id)
);