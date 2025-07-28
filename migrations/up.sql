begin;

create table users (
	id bigserial not null,
	bdate text not null check(date_part('year', now()) - date_part('year', to_date(bdate, 'YYYY-MM-DD')) >= 18),
	name text not null,
	phone text not null check(length(phone) >= 10),
	password_hash bytea not null,
	warehouse_id bigint not null check(warehouse_id > 0),

	primary key(id)
);

create table warehouses (
	id bigserial not null,

	primary key(id)
);

alter table users add constraint fk_warehouses foreign key(warehouse_id) references warehouses(id);

commit;