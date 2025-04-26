INSERT INTO accounts.history_records (id, user_id, event_type, message, created_at, updated_at, deleted_at)
VALUES (DEFAULT, '4c216ab0-a91a-413f-8e97-a32eee7a4ef4'::varchar(255), 'issued'::text, 'Credential is issued'::text,
        '2024-02-29 15:11:09.008043'::timestamp, '2024-02-29 15:11:09.008043'::timestamp, null::timestamp);
INSERT INTO accounts.backups (id, user_id, credentials, created_at, updated_at, deleted_at)
VALUES (DEFAULT, '4c216ab0-a91a-413f-8e97-a32eee7a4ef4'::varchar(255), 'fbhbshbcshbchsbchsv='::bytea,
        '2024-02-29 15:11:09.008043'::timestamp, '2024-02-29 15:11:09.008043'::timestamp, null::timestamp);