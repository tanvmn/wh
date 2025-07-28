begin;

alter table users drop constraint if exists fk_warehouses;

drop table users;
drop table warehouses;

commit;