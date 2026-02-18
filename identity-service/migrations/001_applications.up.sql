create table applications (
                              id          uuid primary key default gen_random_uuid(),
                              code        varchar(100) not null unique,
                              name        varchar(255) not null,
                              active      boolean not null default true,
                              created_at  timestamptz not null default now()
);