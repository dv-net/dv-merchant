insert into aml_services (id, slug, created_at)
values ('621f3ad0-a3b6-4413-8379-f245bbb3e345', 'aml_bot', now()),
       ('63592c9e-8b40-4de5-9359-9a734345996f', 'bit_ok', now())
ON CONFLICT DO NOTHING ;

INSERT INTO aml_service_keys (id, service_id, name, description, created_at)
VALUES
--     AML_BOT
    ('2b5a61f2-9dfc-4abb-969c-cc4ba95d5331', '63592c9e-8b40-4de5-9359-9a734345996f', 'access_key_id', 'Access Key ID', now()),
    ('8b1f1343-0be2-418b-b8a0-44299386d501', '63592c9e-8b40-4de5-9359-9a734345996f', 'secret_key', 'Secret key', now()),

--     BIT_OK
    ('25233cdd-8062-418d-9141-b80b27ff8237', '621f3ad0-a3b6-4413-8379-f245bbb3e345', 'access_id', 'Access ID', now()),
    ('4630e77b-e95e-4ea7-9183-e3e70f1be4d5', '621f3ad0-a3b6-4413-8379-f245bbb3e345', 'access_key', 'Access Key', now())
ON CONFLICT DO NOTHING;