SELECT 'CREATE DATABASE marketflow_db
        WITH OWNER     = postgres
             ENCODING  = ''UTF8''
             TEMPLATE  = template0
             LC_COLLATE = ''en_US.UTF-8''
             LC_CTYPE   = ''en_US.UTF-8'''
WHERE NOT EXISTS (
    SELECT FROM pg_database WHERE datname = 'marketflow_db'
)
\gexec