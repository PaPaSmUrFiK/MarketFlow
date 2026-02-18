create table user_roles
(
    user_id uuid not null references users (id) on delete cascade,
    role_id uuid not null references roles (id) on delete cascade,
    primary key (user_id, role_id)
);

create index idx_user_roles_user on user_roles (user_id);