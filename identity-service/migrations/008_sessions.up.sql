create table sessions (
                          id          uuid primary key default gen_random_uuid(),
                          user_id     uuid not null references users(id) on delete cascade,
                          app_id      uuid not null references applications(id) on delete cascade,
                          user_agent  text,
                          ip_address  inet,
                          created_at  timestamptz not null default now(),
                          revoked     boolean not null default false
);

create index idx_sessions_user on sessions(user_id);
create index idx_sessions_app on sessions(app_id);
create index idx_sessions_active on sessions(user_id) where revoked = false;

alter table refresh_tokens
    add column session_id uuid;

alter table refresh_tokens
    add constraint fk_refresh_session
        foreign key (session_id)
            references sessions(id)
            on delete cascade;

create index idx_refresh_session on refresh_tokens(session_id);