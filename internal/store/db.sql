create table users
(
    id       uuid not null
        primary key,
    login    text not null
        constraint uq_login
            unique,
    password text not null
);

alter table users
    owner to postgres;

create index idx_users_login
    on users (login);

grant delete, insert, select, update on users to go_service_user;

create table games
(
    id              uuid                                             not null
        primary key,
    board           jsonb,
    status          text                     default 'CREATED'::text not null,
    type            text                     default 'PVE'::text     not null,
    move_player_id  uuid,
    owner_player_id uuid,
    created_at      timestamp with time zone default now()           not null
);

alter table games
    owner to postgres;

create table users_games
(
    user_id uuid                       not null
        constraint fk_users_games_users_id
            references users,
    game_id uuid                       not null
        constraint fk_users_games_games_id
            references games,
    role    text default 'OWNER'::text not null,
    primary key (user_id, game_id)
);

alter table users_games
    owner to postgres;

create index idx_users_games_user_id
    on users_games (user_id);

grant delete, insert, select, update on users_games to go_service_user;

create index idx_games_id
    on games (id);

grant delete, insert, select, update on games to go_service_user;

create table refresh_tokens
(
    token   text not null,
    user_id text not null
        constraint uq_user_id_token
            primary key
);

alter table refresh_tokens
    owner to postgres;

grant delete, insert, select, update on refresh_tokens to go_service_user;

