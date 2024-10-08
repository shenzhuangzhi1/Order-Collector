CREATE EXTENSION IF NOT EXISTS "pgcrypto";

drop table if exists oxk_pepe_spot;
create table if not exists oxk_pepe_spot
(
    id uuid primary key default gen_random_uuid(),
    created_at timestamp with time zone default now(),

);

