insert into exchange_keys (exchange_id, name, title, created_at, updated_at)
values
    ('0b08e649-b009-47c9-87c8-4dff132f5362', 'access_key', 'Access key', now(), now()),
    ('0b08e649-b009-47c9-87c8-4dff132f5362', 'secret_key', 'Secret key', now(), now()),
    ('4975c77c-7591-4943-8f02-e6ad62172dab', 'pass_phrase', 'Pass phrase', now(), now()),
    ('4975c77c-7591-4943-8f02-e6ad62172dab', 'secret_key', 'Secret key', now(), now()),
    ('4975c77c-7591-4943-8f02-e6ad62172dab', 'api_key', 'Api key', now(), now()),
    ('9184d445-8eee-4020-916f-1c0c1cbff181', 'secret_key', 'Secret key', now(), now()),
    ('9184d445-8eee-4020-916f-1c0c1cbff181', 'api_key', 'Api key', now(), now()),
    ('06fe1230-4689-4d33-909b-a9c15af64838', 'pass_phrase', 'Pass phrase', now(), now()),
    ('06fe1230-4689-4d33-909b-a9c15af64838', 'secret_key', 'Secret key', now(), now()),
    ('06fe1230-4689-4d33-909b-a9c15af64838', 'access_key', 'Access key', now(), now()),
    ('2e07b084-ff69-4c09-bc9d-e93578260f67', 'secret_key', 'Secret key', now(), now()),
    ('2e07b084-ff69-4c09-bc9d-e93578260f67', 'access_key', 'Access key', now(), now()),
    ('2e07b084-ff69-4c09-bc9d-e93578260f67', 'pass_phrase', 'Pass phrase', now(), now()),
    ('b868db0c-4d51-4842-ab6b-677f62c454fb', 'access_key', 'Access key', now(), now()),
    ('b868db0c-4d51-4842-ab6b-677f62c454fb', 'secret_key', 'Secret key', now(), now()),
    ('2495d8a2-435d-4031-8930-10e82aad9c64', 'secret_key', 'Secret key', now(), now()),
    ('2495d8a2-435d-4031-8930-10e82aad9c64', 'access_key', 'Access key', now(), now())
ON CONFLICT DO NOTHING;