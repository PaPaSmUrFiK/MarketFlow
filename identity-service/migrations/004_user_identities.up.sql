create table user_identities (
                                 id               uuid primary key default gen_random_uuid(),
                                 user_id          uuid not null
                                     references users(id) on delete cascade,
                                 provider         varchar(50) not null,
                                 provider_user_id varchar(255) not null,
                                 created_at       timestamptz not null default now(),
                                 unique (provider, provider_user_id),
                                 unique (user_id, provider)
);

create index idx_user_identities_user on user_identities(user_id);