create table users (
                       id          uuid primary key default gen_random_uuid(),
                       status      varchar(20) not null check (status in ('ACTIVE', 'BLOCKED', 'DELETED')),
                       created_at  timestamptz not null default now(),
                       updated_at  timestamptz not null default now()
);

create index idx_users_status on users(status);
