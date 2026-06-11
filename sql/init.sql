drop database if exists wh;
create database wh;

drop role if exists wu;
create role wu with login password 'pa55word';

alter database wh owner to wu;

create extension if not exists citext;
