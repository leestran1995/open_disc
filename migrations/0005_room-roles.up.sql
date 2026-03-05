create table open_discord.room_roles (
    room_id uuid not null references open_discord.rooms (id) on delete cascade,
    role_id uuid not null references open_discord.roles (id) on delete cascade,
    primary key (room_id, role_id));
    r