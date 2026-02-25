create table open_discord.user_room_stars
(
    user_id uuid not null references open_discord.users,
    room_id uuid not null references open_discord.rooms,
    primary key (user_id, room_id)
);