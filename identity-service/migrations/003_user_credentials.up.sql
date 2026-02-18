create table user_credentials
(
    user_id               uuid primary key
        references users (id) on delete cascade,
    email                 varchar(255) not null unique,
    password_hash         varchar(255),
    email_verified        boolean      not null default false,
    created_at            timestamptz  not null default now(),
    last_password_change  timestamptz,
    failed_login_attempts int                   default 0,
    locked_until          timestamptz
);

create index idx_user_credentials_email on user_credentials (email);