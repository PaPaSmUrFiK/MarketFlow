BEGIN;

DO $$
    DECLARE
        app_uuid uuid;
    BEGIN
        SELECT id INTO app_uuid
        FROM applications
        WHERE code = 'marketflow-web';

        IF app_uuid IS NOT NULL THEN
            DELETE FROM roles
            WHERE app_id = app_uuid
              AND code IN ('USER', 'ADMIN');
        END IF;
    END $$;

DELETE FROM applications WHERE code = 'marketflow-web';

COMMIT;