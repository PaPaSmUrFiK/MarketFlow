BEGIN;

-- Создаём приложение marketflow-web
INSERT INTO applications (code, name)
VALUES ('marketflow-web', 'MarketFlow Web')
ON CONFLICT (code) DO NOTHING;

DO $$
    DECLARE
        app_uuid uuid;
    BEGIN
        SELECT id INTO app_uuid
        FROM applications
        WHERE code = 'marketflow-web';

        -- Создаём базовые роли для приложения
        INSERT INTO roles (app_id, code, description)
        VALUES
            (app_uuid, 'USER',  'Default user role'),
            (app_uuid, 'ADMIN', 'Administrator role')
        ON CONFLICT (app_id, code) DO NOTHING;

        RAISE NOTICE 'Seeded app (ID: %) with USER and ADMIN roles', app_uuid;
    END $$;

COMMIT;