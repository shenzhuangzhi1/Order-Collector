create database oxk_data;

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

drop table if exists oxk_pepe_spot;

create table if not exists oxk_pepe_spot
(
    message    text,
    created_at timestamp with time zone default now()
);