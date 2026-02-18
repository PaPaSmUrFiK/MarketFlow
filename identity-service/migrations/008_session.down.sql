drop index if exists idx_refresh_session;

alter table refresh_tokens
drop constraint if exists fk_refresh_session;

alter table refresh_tokens
drop column if exists session_id;

drop table if exists sessions cascade;