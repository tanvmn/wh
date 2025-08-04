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
	id bigserial not null,
	name text not null,
	address text not null,
	phone text not null,
	email text not null
);

create table if not exists transfer (
	id bigserial not null,
	export_warehouse bigint not null,
	receive_warehouse bigint not null,
	account_id bigint not null,
	expected_dtime timestamp not null
);

create table if not exists bin (
	id bigserial not null,
	warehouse_id bigint,
	capacity real not null
);

create table if not exists tote (
	id bigserial not null,
	warehouse_id bigint,
	capacity real not null
);

create table if not exists inventory (
	id bigserial not null,
	dtime timestamp not null,
	balanced boolean,
	warehouse_id bigint not null,
	account_id bigint not null
);

create table if not exists inventory_item (
	inventory_id bigint not null,
	gtin text not null,
	expected_quantity bigint not null,
	counted_quantity bigint not null
);

create table if not exists account (
	id bigserial not null,
	role text not null,
	bdate date not null check(date_part('year', now()) - date_part('year', bdate) >= 18),
	name text not null,
	phone text not null unique check(length(phone) >= 10),
	password_hash bytea not null,
	warehouse_id bigint default null,
	store_id bigint default null
);

create table if not exists store (
	id bigserial not null,
	name text not null,
	address text not null,
	phone text not null,
	email text not null,
	warehouse_id bigint not null
);

-- drop type if exists status cascade;
-- create type status as enum ('Chờ phản hồi', 'Chờ nhập', 'Đang nhập', 'Kết thúc', 'Từ chối');

create table if not exists purchase (
	id bigserial not null,
	warehouse_id bigint not null,
	account_id bigint not null,
	supplier_id bigint not null,
	expected_dtime timestamp not null,
	status text not null
);

create table if not exists purchase_item (
	purchase_id bigint not null,
	gtin text not null,
	quantity bigint not null
);

-- drop type if exists material cascade;
-- create type material as enum ('Cotton', 'Linen', 'Polyester');

-- drop type if exists size cascade;
-- create type size as enum ('S', 'M', 'L', 'XL', 'XXL');

-- drop type if exists color cascade;
-- create type color as enum ('Đỏ', 'Cam', 'Vàng', 'Lục', 'Lam', 'Chàm', 'Tím', 'Đen', 'Nâu', 'Xám', 'Trắng');

create table if not exists item (
	gtin text not null,
	characteristic text not null,
	volume real not null,
	weight real not null,
	brand text not null,
	material text not null,
	color text not null,
	size text not null,
	price real not null,
	type_id bigint not null
);

create table if not exists type (
	id bigserial not null,
	name text not null
);

create table if not exists resupply (
	id bigserial not null,
	account_id int not null,
	store_id int not null,
	expected_dtime timestamp not null,
	status text not null
);

create table if not exists supplier (
	id bigserial not null,
	name text not null,
	address text not null,
	phone text not null,
	email text not null
);

create table if not exists supplier_item (
	supplier_id bigint not null,
	gtin text not null
);

create table if not exists receive_item (
	purchase_id bigint not null,
	gtin text not null,
	receive_id bigint not null,
	quantity bigint not null
);

create table if not exists receive (
	id bigserial not null,
	purchase_id bigint not null,
	account_id bigint not null,
	expected_dtime timestamp not null,
	actual_dtime timestamp not null default '1000-01-01 00:00:00',
	transfer_id bigint
);

create table if not exists export (
	id bigserial not null,
	resupply_id bigint not null,
	expected_dtime timestamp not null,
	actual_dtime timestamp not null default '1000-01-01 00:00:00',
	transfer_id bigint
);

create table if not exists export_item (
	export_id bigint not null,
	resupply_id bigint not null,
	gtin text not null,
	quantity bigint not null
);

create table if not exists resupply_item (
	resupply_id bigint not null,
	gtin text not null,
	quantity bigint not null
);

create table if not exists seri (
	id text unique not null,
	receive_tote bigint not null,
	pick_tote bigint not null,
	bin_id bigint not null,
	receive_id bigint not null,
	purchase_id bigint not null,
	gtin text not null,
	export_id bigint not null
);

insert into warehouse (name, address, phone, email) values 
('HCM K1', 'địa chỉ HCM K1', '0000000010', 'tan.nguyen2220022@hcmut.edu.vn'),
('DNai K1', 'địa chỉ DNai K1', '0000000011', 'tan.nguyen2220022@hcmut.edu.vn')
;

insert into store (name, address, phone, email, warehouse_id) values
('HCM S1', 'địa chỉ HCM S1', '000000100', 'tan.nguyen2220022@hcmut.edu.vn', 1),
('DNai S1', 'địa chỉ DNai S1', '000000101', 'tan.nguyen2220022@hcmut.edu.vn', 2)
;

insert into tote (warehouse_id, capacity) values 
(1, 92030),
(1, 92030),
(1, 92030),
(1, 92030),
(1, 92030),
(1, 92030),
(1, 92030),
(1, 92030),
(1, 92030),
(1, 92030),
(1, 92030),
(1, 92030),
(1, 92030),
(1, 92030),
(1, 92030),
(1, 92030),
(1, 92030),
(1, 92030),
(1, 92030),
(1, 92030)
;

insert into bin (warehouse_id, capacity) values 
(1, 92030),
(1, 92030),
(1, 92030),
(1, 92030),
(1, 92030),
(1, 92030),
(1, 92030),
(1, 92030),
(1, 92030),
(1, 92030),
(1, 92030),
(1, 92030),
(1, 92030),
(1, 92030),
(1, 92030),
(1, 92030),
(1, 92030),
(1, 92030),
(1, 92030),
(1, 92030)
;

insert into supplier (name, address, phone, email) values
('NCC 1', 'địa chỉ NCC 1', '000001000', 'tanNguyen2220022.hcmut.edu.vn'),
('NCC 2', 'địa chỉ NCC 2', '000001001', 'tanNguyen2220022.hcmut.edu.vn')
;

insert into type (name) values
('Áo sơmi'),
('Áo thun'),
('Áo len'),
('Áo jean'),
('Áo khoác'),
('Quần tây'),
('Quần jean'),
('Quần thun'),
('Váy'),
('Đầm'),
('Onesie')
;

insert into item (gtin, characteristic, volume, weight, brand, material, color, size, price, type_id) values
('4983435734503', 'có túi', 1392, 200, 'GAP', 'Polyester', 'Lam', 'L', 200000, 6),
('8936040400574', 'có túi', 1392, 200, 'Navy', 'Polyester', 'Lam', 'XL', 200000, 8),
('8888021200126', 'có cổ, tay ngắn', 1392, 200, 'Viettien', 'Cotton', 'Đỏ', 'M', 150000, 1),
('4983435764166', 'có cổ, tay ngắn', 1392, 200, 'Gucci', 'Cotton', 'Đỏ', 'XL', 150000, 2),
('4983435734909', 'có cổ, tay dài', 1392, 200, 'Pierre', 'Cotton', 'Đỏ', 'L', 150000, 2)
;

insert into supplier_item (supplier_id, gtin) values
(1, '4983435734503'),
(1, '8936040400574'),
(1, '8888021200126'),
(2, '4983435764166'),
(2, '4983435734909')
;

insert into account (role, bdate, name, phone, password_hash, warehouse_id) values
('Admin', date 'now()' - interval '19 years', 'tim', '0000000001', '$2a$12$sPDZCZEc01jKDxKNDhZgquKZH.4R0TtMn/9sCdnE0OJnrMMcnXPJy', null),
('Kế toán trưởng', date 'now()' - interval '19 years', 'ktt', '0000000002', '$2a$12$sPDZCZEc01jKDxKNDhZgquKZH.4R0TtMn/9sCdnE0OJnrMMcnXPJy', null),
('Thủ kho', date 'now()' - interval '19 years', 'tk', '0000000003', '$2a$12$sPDZCZEc01jKDxKNDhZgquKZH.4R0TtMn/9sCdnE0OJnrMMcnXPJy', 1),
('Kế toán', date 'now()' - interval '19 years', 'kt', '0000000004', '$2a$12$sPDZCZEc01jKDxKNDhZgquKZH.4R0TtMn/9sCdnE0OJnrMMcnXPJy', 1),
('Nhân viên', date 'now()' - interval '19 years', 'nv', '0000000005', '$2a$12$sPDZCZEc01jKDxKNDhZgquKZH.4R0TtMn/9sCdnE0OJnrMMcnXPJy', 1)
;


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

alter table inventory add constraint fk_inventory_warehouse_id foreign key(warehouse_id) references warehouse(id);
alter table inventory add constraint fk_inventory_account_id foreign key(account_id) references account(id);

alter table account add constraint fk_account_warehouse_id foreign key(warehouse_id) references warehouse(id);
alter table account add constraint fk_account_store_id foreign key(store_id) references store(id);

alter table store add constraint fk_store_warehouse_id foreign key(warehouse_id) references warehouse(id);

alter table tote add constraint fk_tote_store_warehouse_id foreign key(warehouse_id) references warehouse(id);
alter table bin add constraint fk_bin_store_warehouse_id foreign key(warehouse_id) references warehouse(id);

alter table purchase add constraint fk_purchase_warehouse_id foreign key(warehouse_id) references warehouse(id);
alter table purchase add constraint fk_purchase_supplier_id foreign key(supplier_id) references supplier(id);
alter table purchase add constraint fk_purchase_account_id foreign key(account_id) references account(id);

alter table purchase_item add constraint fk_purchase_item_purchase_id foreign key(purchase_id) references purchase(id);
alter table purchase_item add constraint fk_purchase_item_gtin foreign key(gtin) references item(gtin);

alter table item add constraint fk_item_type_id foreign key(type_id) references type(id);

alter table resupply add constraint fk_resupply_account_id foreign key(account_id) references account(id);
alter table resupply add constraint fk_resupply_store_id foreign key(store_id) references store(id);

alter table supplier_item add constraint fk_supplier_gtin foreign key(gtin) references item(gtin);
alter table supplier_item add constraint fk_supplier_supplier_id foreign key(supplier_id) references supplier(id);

alter table receive_item add constraint fk_receive_item_receive_id foreign key(receive_id) references receive(id);
alter table receive_item add constraint fk_receive_item_purchase_id_gtin foreign key(purchase_id, gtin) references purchase_item(purchase_id, gtin);

alter table receive add constraint fk_receive_purchase_id foreign key(purchase_id) references purchase(id);
alter table receive add constraint fk_receive_account_id foreign key(account_id) references account(id);
alter table receive add constraint fk_receive_transfer_id foreign key(transfer_id) references transfer(id);

alter table export add constraint fk_export_resupply_id foreign key(resupply_id) references resupply(id);
alter table export add constraint fk_export_transfer_id foreign key(transfer_id) references transfer(id);

alter table export_item add constraint fk_export_item_export_id foreign key(export_id) references export(id);
alter table export_item add constraint fk_export_item_resupply_id_gtin foreign key(resupply_id, gtin) references resupply_item(resupply_id, gtin);

alter table resupply_item add constraint fk_resupply_item_resupply_id foreign key(resupply_id) references resupply(id);
alter table resupply_item add constraint fk_resupply_item_gtin foreign key(gtin) references item(gtin);

alter table seri add constraint fk_seri_receive_tote foreign key(receive_tote) references tote(id);
alter table seri add constraint fk_seri_pick_tote foreign key(pick_tote) references tote(id);
alter table seri add constraint fk_seri_bin_id foreign key(bin_id) references bin(id);
alter table seri add constraint fk_seri_receive_id_purchase_id_gtin foreign key(receive_id, purchase_id, gtin) references receive_item(purchase_id, receive_id, gtin);
alter table seri add constraint fk_seri_export_id foreign key(export_id) references export(id);

CREATE TABLE sessions (
	token TEXT PRIMARY KEY,
	data BYTEA NOT NULL,
	expiry TIMESTAMPTZ NOT NULL
);

CREATE INDEX sessions_expiry_idx ON sessions (expiry);

commit;