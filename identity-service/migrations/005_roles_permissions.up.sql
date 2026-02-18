create table roles
(
    id          uuid primary key default gen_random_uuid(),
    app_id      uuid         not null references applications (id) on delete cascade,
    code        varchar(100) not null,
    description varchar(255),
    unique (app_id, code)
);

create index idx_roles_app on roles (app_id);

create table permissions
(
    id          uuid primary key default gen_random_uuid(),
    app_id      uuid         not null references applications (id) on delete cascade,
    code        varchar(150) not null,
    description varchar(255),
    unique (app_id, code)
);

create index idx_permissions_app on permissions (app_id);

create table role_permissions
(
    role_id       uuid not null references roles (id) on delete cascade,
    permission_id uuid not null references permissions (id) on delete cascade,
    primary key (role_id, permission_id)
);

create index idx_role_permissions_role on role_permissions (role_id);