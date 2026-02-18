create table refresh_tokens
(
    id         uuid primary key      default gen_random_uuid(),
    user_id    uuid         not null references users (id) on delete cascade,
    app_id     uuid         not null references applications (id) on delete cascade,
    token_hash varchar(255) not null unique,
    expires_at timestamptz  not null,
    revoked    boolean      not null default false,
    created_at timestamptz  not null default now()
);

create index idx_refresh_user on refresh_tokens (user_id);
create index idx_refresh_active on refresh_tokens (user_id) where revoked = false;