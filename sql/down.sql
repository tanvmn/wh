begin;

alter table account drop constraint if exists fk_warehouse;

drop table account;
drop table warehouse;

commit;
