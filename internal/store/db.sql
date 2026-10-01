create database tic_tac_toe
    with owner postgres;

grant connect on database tic_tac_toe to go_service_user;

create table public.users
(
    id       uuid not null
        primary key,
    login    text not null
        constraint uq_login
            unique,
    password text not null
);

alter table public.users
    owner to postgres;

create index idx_users_login
    on public.users (login);

grant delete, insert, select, update on public.users to go_service_user;

create table public.games
(
    id              uuid                         not null
        primary key,
    board           jsonb,
    status          text default 'CREATED'::text not null,
    type            text default 'PVE'::text     not null,
    move_player_id  text,
    owner_player_id text default 0               not null
);

alter table public.games
    owner to postgres;

create table public.users_games
(
    user_id uuid                       not null
        constraint fk_users_games_users_id
            references public.users,
    game_id uuid                       not null
        constraint fk_users_games_games_id
            references public.games,
    role    text default 'OWNER'::text not null,
    primary key (user_id, game_id)
);

alter table public.users_games
    owner to postgres;

create index idx_users_games_user_id
    on public.users_games (user_id);

grant delete, insert, select, update on public.users_games to go_service_user;

create index idx_games_id
    on public.games (id);

grant delete, insert, select, update on public.games to go_service_user;

