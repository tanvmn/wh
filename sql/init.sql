drop database if exists wh;
create database wh;

drop role if exists wu;
create role wu with login password 'pa55word';

create extension if not exists citext;
