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

insert into open_discord.roles (name) values ('admin');
insert into open_discord.roles (name) values ('default');

CREATE OR REPLACE FUNCTION open_discord.prevent_protected_role_deletion()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.name IN ('admin', 'default') THEN
        RAISE EXCEPTION 'Cannot delete protected role: %', OLD.name;
    END IF;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER protect_system_roles
    BEFORE DELETE ON open_discord.roles
    FOR EACH ROW
    EXECUTE FUNCTION open_discord.prevent_protected_role_deletion();

CREATE OR REPLACE FUNCTION open_discord.prevent_protected_role_rename()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.name IN ('admin', 'default') AND NEW.name != OLD.name THEN
        RAISE EXCEPTION 'Cannot rename protected role: %', OLD.name;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER protect_system_role_names
    BEFORE UPDATE ON open_discord.roles
    FOR EACH ROW
    EXECUTE FUNCTION open_discord.prevent_protected_role_rename();