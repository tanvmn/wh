begin;

-- alter table warehouse drop constraint if exists pk_warehouse cascade;
-- alter table transfer drop constraint if exists pk_transfer cascade;
-- alter table inventory drop constraint if exists pk_inventory cascade;
-- alter table inventory_item drop constraint if exists pk_inventory_item cascade;
-- alter table account drop constraint if exists pk_account cascade;
-- alter table store drop constraint if exists pk_store cascade;
-- alter table purchase drop constraint if exists pk_purchase cascade;
-- alter table purchase_item drop constraint if exists pk_purchase_item cascade;
-- alter table item drop constraint if exists pk_item cascade;
-- alter table type drop constraint if exists pk_type cascade;
-- alter table resupply drop constraint if exists pk_resupply cascade;
-- alter table tote drop constraint if exists pk_tote cascade;
-- alter table bin drop constraint if exists pk_bin cascade;
-- alter table supplier drop constraint if exists pk_supplier cascade;
-- alter table supplier_item drop constraint if exists pk_supplier_item cascade;
-- alter table receive_item drop constraint if exists pk_receive_item cascade;
-- alter table receive drop constraint if exists pk_receive cascade;
-- alter table export drop constraint if exists pk_export cascade;
-- alter table export_item drop constraint if exists pk_export_item cascade;
-- alter table resupply_item drop constraint if exists pk_resupply_item cascade;
-- alter table seri drop constraint if exists pk_seri cascade;

-- alter table transfer drop constraint if exists fk_transfer_export_warehouse cascade;
-- alter table transfer drop constraint if exists fk_transfer_receive_warehouse cascade;
-- alter table transfer drop constraint if exists fk_transfer_account_id cascade;

-- alter table inventory drop constraint if exists fk_inventory_warehouse_id cascade;
-- alter table inventory drop constraint if exists fk_inventory_account_id cascade;

-- alter table inventory_item drop constraint if exists fk_inventory_item_inventory_id cascade;
-- alter table inventory_item drop constraint if exists fk_inventory_gtin cascade;

-- alter table account drop constraint if exists fk_account_warehouse_id cascade;
-- alter table account drop constraint if exists fk_account_store_id cascade;

drop table if exists warehouse cascade;
drop table if exists transfer cascade;
drop table if exists bin cascade;
drop table if exists tote cascade;
drop table if exists inventory cascade;
drop table if exists inventory_item cascade;
drop table if exists account cascade;
drop table if exists store cascade;
drop table if exists purchase cascade;
drop table if exists purchase_item cascade;
drop table if exists item cascade;
drop table if exists type cascade;
drop table if exists resupply cascade;
drop table if exists supplier cascade;
drop table if exists supplier_item cascade;
drop table if exists receive_item cascade;
drop table if exists receive cascade;
drop table if exists export cascade;
drop table if exists export_item cascade;
drop table if exists resupply_item cascade;
drop table if exists seri cascade;
drop table if exists sessions cascade;


create table if not exists warehouse (
	id serial not null,
	name text not null,
	address text not null,
	phone text not null,
	email text not null
);

create table if not exists transfer (
	id serial not null,
	export_warehouse integer not null,
	receive_warehouse integer not null,
	account_id integer not null,
	expected_dtime timestamp not null
);

create table if not exists bin (
	id serial not null,
	warehouse_id integer,
	capacity real not null
);

create table if not exists tote (
	id serial not null,
	warehouse_id integer,
	capacity real not null
);

create table if not exists inventory (
	id serial not null,
	dtime timestamp not null,
	balanced boolean,
	warehouse_id integer not null,
	account_id integer not null
);

create table if not exists inventory_item (
	inventory_id integer not null,
	gtin text not null,
	expected_quantity integer not null,
	counted_quantity integer not null
);

create table if not exists account (
	id serial not null,
	role text not null,
	bdate date not null check(date_part('year', now()) - date_part('year', bdate) >= 18),
	name text not null,
	phone text not null unique check(length(phone) >= 10),
	password_hash bytea not null,
	warehouse_id integer default null,
	store_id integer default null
);

create table if not exists store (
	id integer not null,
	name text not null,
	address text not null,
	phone text not null,
	email text not null,
	warehouse_id integer not null
);

drop type if exists status cascade;
create type status as enum ('awaiting response', 'awaiting receive', 'receiving', 'ended', 'denied');

create table if not exists purchase (
	id serial not null,
	warehouse_id integer not null,
	account_id integer not null,
	supplier_id integer not null,
	expected_dtime timestamp not null,
	status status not null
);

create table if not exists purchase_item (
	purchase_id integer not null,
	gtin text not null,
	quantity integer not null
);

create table if not exists item (
	gtin text not null,
	characteristic text not null,
	length real not null,
	width real not null,
	height real not null,
	weight real not null,
	brand text not null,
	material text not null,
	color text not null,
	size text not null,
	price real not null,
	type_id text not null
);

create table if not exists type (
	id serial not null,
	name text not null
);

create table if not exists resupply (
	id serial not null,
	account_id int not null,
	store_id int not null,
	expected_dtime timestamp not null,
	status status not null
);

create table if not exists supplier (
	id serial not null,
	name text not null,
	address text not null,
	phone text not null,
	email text not null
);

create table if not exists supplier_item (
	supplier_id integer not null,
	gtin text not null
);

create table if not exists receive_item (
	purchase_id integer not null,
	gtin text not null,
	receive_id integer not null,
	quantity integer not null
);

create table if not exists receive (
	id serial not null,
	purchase_id integer not null,
	account_id integer not null,
	expected_dtime timestamp not null,
	actual_dtime timestamp not null default '1000-01-01 00:00:00',
	transfer_id integer
);

create table if not exists export (
	id serial not null,
	resupply_id integer not null,
	expected_dtime timestamp not null,
	actual_dtime timestamp not null default '1000-01-01 00:00:00',
	transfer_id integer
);

create table if not exists export_item (
	export_id integer not null,
	resupply_id integer not null,
	gtin text not null,
	quantity integer not null
);

create table if not exists resupply_item (
	resupply_id integer not null,
	gtin text not null,
	quantity integer not null
);

create table if not exists seri (
	id text unique not null,
	receive_tote integer not null,
	pick_tote integer not null,
	bin_id integer not null,
	receive_id integer not null,
	purchase_id integer not null,
	gtin text not null,
	expected_id integer not null
);


alter table warehouse add constraint pk_warehouse primary key(id);
alter table transfer add constraint pk_transfer primary key(id);
alter table inventory add constraint pk_inventory primary key(id);
alter table inventory_item add constraint pk_inventory_item primary key(inventory_id, gtin);
alter table account add constraint pk_account primary key(id);
alter table store add constraint pk_store primary key(id);
alter table purchase add constraint pk_purchase primary key(id);
alter table purchase_item add constraint pk_purchase_item primary key(purchase_id, gtin);
alter table item add constraint pk_item primary key(gtin);
alter table type add constraint pk_type primary key(id);
alter table resupply add constraint pk_resupply primary key(id);
alter table tote add constraint pk_tote primary key(id);
alter table bin add constraint pk_bin primary key(id);
alter table supplier add constraint pk_supplier primary key(id);
alter table supplier_item add constraint pk_supplier_item primary key(supplier_id, gtin);
alter table receive_item add constraint pk_receive_item primary key(receive_id, gtin, purchase_id);
alter table receive add constraint pk_receive primary key(id);
alter table export add constraint pk_export primary key(id);
alter table export_item add constraint pk_export_item primary key(export_id, resupply_id, gtin);
alter table resupply_item add constraint pk_resupply_item primary key(resupply_id, gtin);
alter table seri add constraint pk_seri primary key(id);

alter table transfer add constraint fk_transfer_export_warehouse foreign key(export_warehouse) references warehouse(id);
alter table transfer add constraint fk_transfer_receive_warehouse foreign key(receive_warehouse) references warehouse(id);
alter table transfer add constraint fk_transfer_account_id foreign key(account_id) references account(id);

alter table inventory add constraint fk_warehouse_id foreign key(warehouse_id) references warehouse(id);
alter table inventory add constraint fk_account_id foreign key(account_id) references account(id);


CREATE TABLE sessions (
	token TEXT PRIMARY KEY,
	data BYTEA NOT NULL,
	expiry TIMESTAMPTZ NOT NULL
);

CREATE INDEX sessions_expiry_idx ON sessions (expiry);


insert into account (role, bdate, name, phone, password_hash, warehouse_id) values
('Admin', date 'now()' - interval '19 years', 'tim', '0000000001', '$2a$12$sPDZCZEc01jKDxKNDhZgquKZH.4R0TtMn/9sCdnE0OJnrMMcnXPJy', null),
('Kế toán trưởng', date 'now()' - interval '19 years', 'ktt', '0000000002', '$2a$12$sPDZCZEc01jKDxKNDhZgquKZH.4R0TtMn/9sCdnE0OJnrMMcnXPJy', null),
('Thủ kho', date 'now()' - interval '19 years', 'tk', '0000000003', '$2a$12$sPDZCZEc01jKDxKNDhZgquKZH.4R0TtMn/9sCdnE0OJnrMMcnXPJy', 1),
('Kế toán', date 'now()' - interval '19 years', 'kt', '0000000004', '$2a$12$sPDZCZEc01jKDxKNDhZgquKZH.4R0TtMn/9sCdnE0OJnrMMcnXPJy', 1),
('Nhân viên', date 'now()' - interval '19 years', 'nv', '0000000005', '$2a$12$sPDZCZEc01jKDxKNDhZgquKZH.4R0TtMn/9sCdnE0OJnrMMcnXPJy', 1)
;

commit;