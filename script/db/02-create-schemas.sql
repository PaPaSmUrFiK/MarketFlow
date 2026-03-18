\connect marketflow_db

-- Расширения
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Схема Identity Service
CREATE SCHEMA IF NOT EXISTS identity;

COMMENT ON SCHEMA identity IS 'Identity Service — users, sessions, tokens, roles';

-- Создаём роль через \gexec (работает при передаче через stdin)
SELECT 'CREATE ROLE ' || :'identity_user' || ' LOGIN PASSWORD ''' || :'identity_pass' || ''''
WHERE NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = :'identity_user')
\gexec

-- Права на схему identity и public через \gexec
SELECT format('GRANT USAGE, CREATE ON SCHEMA identity TO %I', :'identity_user') \gexec
SELECT format('GRANT USAGE, CREATE ON SCHEMA public TO %I', :'identity_user') \gexec

-- Права на будущие таблицы в identity
SELECT format('ALTER DEFAULT PRIVILEGES IN SCHEMA identity GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO %I', :'identity_user') \gexec
SELECT format('ALTER DEFAULT PRIVILEGES IN SCHEMA identity GRANT USAGE, SELECT ON SEQUENCES TO %I', :'identity_user') \gexec

-- Права на будущие таблицы в public (только schema_migrations)
SELECT format('ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO %I', :'identity_user') \gexec

-- search_path: identity первый — CREATE TABLE без префикса идут туда
SELECT format('ALTER ROLE %I SET search_path = identity, public', :'identity_user') \gexec
