begin;

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
drop table if exists serial cascade;
drop table if exists difference cascade;
drop table if exists sessions cascade;
drop table if exists difference_serial cascade;
drop table if exists package cascade;
drop table if exists package_item cascade;
drop table if exists package_serial cascade;
drop table if exists inventory cascade;
drop table if exists inventory_serial cascade;


create table if not exists warehouse (
	id bigserial not null,
	name text not null,
	address text not null,
	phone text not null,
	version integer not null default 1,
	email text not null
);

create table if not exists transfer (
	id bigserial not null,
	export_warehouse bigint not null,
	receive_warehouse bigint not null,
	account_id bigint not null,
	created_at timestamp not null default now(),
	version integer not null default 1,
	note text default 'none',
	expected_at timestamp not null
);

create table if not exists bin (
	id bigserial not null,
	warehouse_id bigint not null,
	shelf bigint not null,
	row bigint not null,
	col bigint not null,
	version integer not null default 1,
	capacity real not null default 37480
);

create table if not exists tote (
	id bigserial not null,
	warehouse_id bigint not null,
	version integer not null default 1,
	capacity real not null default 37480
);

create table if not exists inventory (
	id bigserial not null,
	created_at timestamp not null default now(),
	expected_at timestamp not null,
	started_at timestamp not null default '1000-01-01',
	ended_at timestamp not null default '1000-01-01',
	balanced boolean not null default false,
	warehouse_id bigint not null,
	version integer not null default 1,
	note text default 'none',
	account_id bigint not null
);

create table if not exists inventory_serial (
	inventory_id bigint not null
	,serial text not null
	,result text not null default 'unchecked'
	,note text not null default 'none'
);

create table if not exists account (
	id bigserial not null,
	role text not null,
	bdate date not null check(date_part('year', now()) - date_part('year', bdate) >= 18),
	name text not null,
	phone text not null unique check(length(phone) >= 10),
	password_hash bytea not null,
	warehouse_id bigint default null,
	version integer not null default 1,
	store_id bigint default null
);

create table if not exists store (
	id bigserial not null,
	name text not null,
	address text not null,
	phone text not null,
	email text not null,
	version integer not null default 1,
	warehouse_id bigint not null
);

drop type if exists status cascade;
create type status as enum ('CHỜ PHẢN HỒI', 'CHỜ NHẬP', 'ĐANG NHẬP', 'CHỜ XUẤT', 'ĐANG XUẤT', 'KẾT THÚC', 'TỪ CHỐI');

create table if not exists purchase (
	id bigserial not null,
	warehouse_id bigint not null,
	account_id bigint not null,
	supplier_id bigint not null,
	expected_at timestamp not null,
	created_at timestamp not null default now(),
	version integer not null default 1,
	note text default 'none',
	-- receive_add_owner bigint default 0,
	receive_add_owner bigint,
	status status not null default 'CHỜ PHẢN HỒI'
);

create table if not exists purchase_item (
	purchase_id bigint not null,
	gtin text not null,
	version integer not null default 1,
	quantity bigint not null
);

drop type if exists material cascade;
create type material as enum ('Cotton', 'Linen', 'Polyester', 'Silk', 'Wool', 'Rayon', 'Denim');

drop type if exists size cascade;
create type size as enum ('S', 'M', 'L', 'XL', 'XXL');

drop type if exists color cascade;
create type color as enum ('Đỏ', 'Cam', 'Vàng', 'Lục', 'Lam', 'Chàm', 'Tím', 'Đen', 'Nâu', 'Xám', 'Trắng', 'Hồng');

drop type if exists brand cascade;
-- create type brand as enum ('Gucci', 'GAP', 'Navy', 'Viettien', 'Pierre', 'H&M', 'Zara');

drop type if exists item_type cascade;
drop type if exists type cascade;
create type type as enum (
'Áo sơmi',
'Áo thun',
'Áo len',
'Áo jean',
'Áo da',
'Áo khoác',
'Quần jean',
'Quần tây',
'Quần thun',
'Quần da',
'Váy',
'Đầm',
'Onesie'
);

create table if not exists item (
	gtin text not null,
	characteristic text not null,
	volume real not null,
	weight bigint not null,
	brand text not null,
	material material not null,
	color color not null,
	size size not null,
	price real not null default -1,
	currency text not null default 'VND',
	shelf_life bigint not null,		-- months
	img_fspath text not null,
	version integer not null default 1,
	type type not null
);

create table if not exists resupply (
	id bigserial not null,
	created_at timestamp not null default now(),
	expected_at timestamp not null,
	status status not null default 'CHỜ PHẢN HỒI',
	note text default 'none',
	account_id int not null,
	store_id int not null,
	export_add_owner bigint,
	version integer not null default 1
);

create table if not exists resupply_item (
	resupply_id bigint not null,
	-- receive_id bigint not null,
	-- purchase_id bigint not null,
	gtin text not null,
	version integer not null default 1,
	quantity bigint not null
);


create table if not exists supplier (
	id bigserial not null,
	name text not null,
	address text not null,
	phone text not null,
	version integer not null default 1,
	email text not null
);

create table if not exists supplier_item (
	supplier_id bigint not null,
	version integer not null default 1,
	gtin text not null
);

create table if not exists receive (
	id bigserial not null,
	purchase_id bigint not null,
	account_id bigint not null,
	processed_by bigint,
	putaway_by bigint,
	created_at timestamp not null default now(),
	expected_at timestamp not null,
	actual_at timestamp not null default '1000-01-01 00:00:00',
	putaway_at timestamp not null default '1000-01-01 00:00:00',
	version integer not null default 1,
	note text default 'none',
	putaway_note text default 'none',
	voucher_id text not null default 'empty',
	transfer_id bigint
);

create table if not exists receive_item (
	purchase_id bigint not null,
	gtin text not null,
	receive_id bigint not null,
	version integer not null default 1,
	note text default 'none',
	putaway_note text default 'none',
	quantity bigint not null
);

create table if not exists export (
	id bigserial not null,
	account_id bigint not null,
	picked_by bigint,
	packed_by bigint,
	created_at timestamp not null default now(),
	expected_at timestamp not null,
	picked_at timestamp not null default '1000-01-01 00:00:00',
	packed_at timestamp not null default '1000-01-01 00:00:00',
	note text default 'none',
	pick_note text default 'none',
	pack_note text default 'none',
	voucher_id text not null default 'empty',
	resupply_id bigint not null,
	transfer_id bigint,
	version integer not null default 1
);

create table if not exists export_item (
	export_id bigint not null,
	resupply_id bigint not null,
	gtin text not null,
	version integer not null default 1,
	note text default 'none',
	pick_note text default 'none',
	pack_note text default 'none',
	quantity bigint not null
);

create table if not exists serial (
	nanoid text unique not null,
	receive_tote bigint not null,
	pick_tote bigint,
	bin_id bigint,
	receive_id bigint not null,
	purchase_id bigint not null,
	gtin text not null,
	export_id bigint,
	resupply_id bigint,
	version integer not null default 1
	-- packed boolean not null default false
	-- resupply_id bigint,
);

create table if not exists difference_serial (
	activity_id text not null,
	nanoid text unique not null,
	receive_tote bigint not null,
	pick_tote bigint,
	bin_id bigint,
	receive_id bigint not null,
	purchase_id bigint not null,
	gtin text not null,
	export_id bigint,
	resupply_id bigint
);

create table if not exists package (
	nanoid text not null
	,export_id bigint not null
);

create table if not exists package_item (
	package_nanoid text not null
	,gtin text not null
	,quantity bigint not null
	,pack_note text not null default 'none'
);

create table if not exists package_serial (
	package_nanoid text not null
	,gtin text not null
	,serial_nanoid text not null
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
(1, 37480),
(1, 37480),
(1, 37480),
(1, 37480),
(1, 37480),
(1, 37480),
(1, 37480),
(1, 37480),
(1, 37480),
(1, 37480),
(1, 37480),
(1, 37480),
(1, 37480),
(1, 37480),
(1, 37480),
(1, 37480),
(1, 37480),
(1, 37480),
(1, 37480),
(1, 37480),
(2, 37480),
(2, 37480),
(2, 37480),
(2, 37480),
(2, 37480),
(2, 37480),
(2, 37480),
(2, 37480),
(2, 37480),
(2, 37480),
(2, 37480),
(2, 37480),
(2, 37480),
(2, 37480),
(2, 37480),
(2, 37480),
(2, 37480),
(2, 37480),
(2, 37480),
(2, 37480)
;

insert into bin (warehouse_id, shelf, row, col, capacity) values 
(1, 1, 1, 1, 37480),
(1, 1, 1, 2, 37480),
(1, 1, 1, 3, 37480),
(1, 1, 1, 4, 37480),
(1, 1, 1, 5, 37480),
(1, 1, 2, 1, 37480),
(1, 1, 2, 2, 37480),
(1, 1, 2, 3, 37480),
(1, 1, 2, 4, 37480),
(1, 1, 2, 5, 37480),
(1, 1, 3, 1, 37480),
(1, 1, 3, 2, 37480),
(1, 1, 3, 3, 37480),
(1, 1, 3, 4, 37480),
(1, 1, 3, 5, 37480),
(1, 1, 4, 1, 37480),
(1, 1, 4, 2, 37480),
(1, 1, 4, 3, 37480),
(1, 1, 4, 4, 37480),
(1, 1, 4, 5, 37480),
(1, 1, 5, 1, 37480),
(1, 1, 5, 2, 37480),
(1, 1, 5, 3, 37480),
(1, 1, 5, 4, 37480),
(1, 1, 5, 5, 37480),

(1, 2, 1, 1, 37480),
(1, 2, 1, 2, 37480),
(1, 2, 1, 3, 37480),
(1, 2, 1, 4, 37480),
(1, 2, 1, 5, 37480),
(1, 2, 2, 1, 37480),
(1, 2, 2, 2, 37480),
(1, 2, 2, 3, 37480),
(1, 2, 2, 4, 37480),
(1, 2, 2, 5, 37480),
(1, 2, 3, 1, 37480),
(1, 2, 3, 2, 37480),
(1, 2, 3, 3, 37480),
(1, 2, 3, 4, 37480),
(1, 2, 3, 5, 37480),
(1, 2, 4, 1, 37480),
(1, 2, 4, 2, 37480),
(1, 2, 4, 3, 37480),
(1, 2, 4, 4, 37480),
(1, 2, 4, 5, 37480),
(1, 2, 5, 1, 37480),
(1, 2, 5, 2, 37480),
(1, 2, 5, 3, 37480),
(1, 2, 5, 4, 37480),
(1, 2, 5, 5, 37480),

(1, 3, 1, 1, 37480),
(1, 3, 1, 2, 37480),
(1, 3, 1, 3, 37480),
(1, 3, 1, 4, 37480),
(1, 3, 1, 5, 37480),
(1, 3, 2, 1, 37480),
(1, 3, 2, 2, 37480),
(1, 3, 2, 3, 37480),
(1, 3, 2, 4, 37480),
(1, 3, 2, 5, 37480),
(1, 3, 3, 1, 37480),
(1, 3, 3, 2, 37480),
(1, 3, 3, 3, 37480),
(1, 3, 3, 4, 37480),
(1, 3, 3, 5, 37480),
(1, 3, 4, 1, 37480),
(1, 3, 4, 2, 37480),
(1, 3, 4, 3, 37480),
(1, 3, 4, 4, 37480),
(1, 3, 4, 5, 37480),
(1, 3, 5, 1, 37480),
(1, 3, 5, 2, 37480),
(1, 3, 5, 3, 37480),
(1, 3, 5, 4, 37480),
(1, 3, 5, 5, 37480),

(2, 1, 1, 1, 37480),
(2, 1, 1, 2, 37480),
(2, 1, 1, 3, 37480),
(2, 1, 1, 4, 37480),
(2, 1, 1, 5, 37480),
(2, 1, 2, 1, 37480),
(2, 1, 2, 2, 37480),
(2, 1, 2, 3, 37480),
(2, 1, 2, 4, 37480),
(2, 1, 2, 5, 37480),
(2, 1, 3, 1, 37480),
(2, 1, 3, 2, 37480),
(2, 1, 3, 3, 37480),
(2, 1, 3, 4, 37480),
(2, 1, 3, 5, 37480),
(2, 1, 4, 1, 37480),
(2, 1, 4, 2, 37480),
(2, 1, 4, 3, 37480),
(2, 1, 4, 4, 37480),
(2, 1, 4, 5, 37480),
(2, 1, 5, 1, 37480),
(2, 1, 5, 2, 37480),
(2, 1, 5, 3, 37480),
(2, 1, 5, 4, 37480),
(2, 1, 5, 5, 37480),

(2, 2, 1, 1, 37480),
(2, 2, 1, 2, 37480),
(2, 2, 1, 3, 37480),
(2, 2, 1, 4, 37480),
(2, 2, 1, 5, 37480),
(2, 2, 2, 1, 37480),
(2, 2, 2, 2, 37480),
(2, 2, 2, 3, 37480),
(2, 2, 2, 4, 37480),
(2, 2, 2, 5, 37480),
(2, 2, 3, 1, 37480),
(2, 2, 3, 2, 37480),
(2, 2, 3, 3, 37480),
(2, 2, 3, 4, 37480),
(2, 2, 3, 5, 37480),
(2, 2, 4, 1, 37480),
(2, 2, 4, 2, 37480),
(2, 2, 4, 3, 37480),
(2, 2, 4, 4, 37480),
(2, 2, 4, 5, 37480),
(2, 2, 5, 1, 37480),
(2, 2, 5, 2, 37480),
(2, 2, 5, 3, 37480),
(2, 2, 5, 4, 37480),
(2, 2, 5, 5, 37480),

(2, 3, 1, 1, 37480),
(2, 3, 1, 2, 37480),
(2, 3, 1, 3, 37480),
(2, 3, 1, 4, 37480),
(2, 3, 1, 5, 37480),
(2, 3, 2, 1, 37480),
(2, 3, 2, 2, 37480),
(2, 3, 2, 3, 37480),
(2, 3, 2, 4, 37480),
(2, 3, 2, 5, 37480),
(2, 3, 3, 1, 37480),
(2, 3, 3, 2, 37480),
(2, 3, 3, 3, 37480),
(2, 3, 3, 4, 37480),
(2, 3, 3, 5, 37480),
(2, 3, 4, 1, 37480),
(2, 3, 4, 2, 37480),
(2, 3, 4, 3, 37480),
(2, 3, 4, 4, 37480),
(2, 3, 4, 5, 37480),
(2, 3, 5, 1, 37480),
(2, 3, 5, 2, 37480),
(2, 3, 5, 3, 37480),
(2, 3, 5, 4, 37480),
(2, 3, 5, 5, 37480)
;

insert into supplier (name, address, phone, email) values
('NCC 1', 'địa chỉ NCC 1', '000001000', 'tan.nguyen2220022@hcmut.edu.vn'),
('NCC 2', 'địa chỉ NCC 2', '000001001', 'tan.nguyen2220022@hcmut.edu.vn')
;

-- insert into item (gtin, characteristic, volume, weight, brand, material, color, size, price, type, shelf_life, img_fspath) values
-- ('4983435734503', 'có túi', 1392, 200, 'GAP', 'Polyester', 'Lam', 200000, 'Quần tây', 7, array ['rec', 'item', 'img', '4983435734503.jpg']),			-- quần tây lam
-- ('8936040400574', 'có túi', 1392, 200, 'Navy', 'Polyester', 'Lục', 'XL', 200000, 'Quần thun', 7, array ['rec', 'item', 'img', '8936040400574.jpeg']),		-- quần thun lục
-- ('8888021200126', 'có cổ, tay dài', 1392, 200, 'Viettien', 'Cotton', 'Trắng', 150000, 'Áo sơmi', 7, array ['rec', 'item', 'img', '8888021200126.jpg']),	-- áo somi trắng tay dài
-- ('4983435764166', 'có cổ, tay ngắn', 1392, 200, 'Gucci', 'Cotton', 'Vàng', 'XL', 150000, 'Áo thun', 7, array ['rec', 'item', 'img', '4983435764166.jpeg']),		-- áo thun vàng tay ngắn
-- ('4983435734909', 'có cổ, tay dài', 1392, 200, 'Pierre', 'Cotton', 'Đen', 150000, 'Áo khoác', 7, array ['rec', 'item', 'img', '4983435734909.jpeg'])		-- áo khoác đen tay dài
-- ;

insert into item (gtin, characteristic, volume, weight, brand, material, color, size, price, type, shelf_life, img_fspath) values
('619659115906', 'có túi', 1392, 200, 'Navy', 'Polyester', 'Lục', 'XL', 200000, 'Quần thun', 7, 'item/img/619659115906.jpeg'),			-- quần thun lục
('8888021200126', 'có cổ, tay dài', 1392, 200, 'Viettien', 'Cotton', 'Trắng', 'S', 150000, 'Áo sơmi', 7, 'item/img/8888021200126.jpeg'),	-- áo somi trắng tay dài
('4983435764166', 'không cổ, tay ngắn', 1392, 200, 'Gucci', 'Cotton', 'Vàng', 'L', 150000, 'Áo thun', 7, 'item/img/4983435764166.jpeg'),	-- áo thun vàng tay ngắn
('8904091104109', 'có túi', 1392, 200, 'GAP', 'Polyester', 'Lam', 'M', 200000, 'Quần tây', 7, 'item/img/8904091104109.jpeg'),			-- quần tây lam
('8936134272407', 'có cổ, tay dài', 1392, 200, 'Pierre', 'Linen', 'Đen', 'XXL', 150000, 'Áo khoác', 7, 'item/img/8936134272407.jpeg')		-- áo khoác đen tay dài
;

insert into supplier_item (supplier_id, gtin) values
(1, '619659115906'),
(1, '8904091104109'),
(1, '8888021200126'),
(2, '4983435764166'),
(2, '8936134272407')
;

insert into account (role, bdate, name, phone, password_hash, warehouse_id) values
('Admin', date 'now()' - interval '19 years', 'admin', '0000000001', '$2a$12$sPDZCZEc01jKDxKNDhZgquKZH.4R0TtMn/9sCdnE0OJnrMMcnXPJy', 1),
('Kế toán trưởng', date 'now()' - interval '19 years', 'ktt', '0000000002', '$2a$12$sPDZCZEc01jKDxKNDhZgquKZH.4R0TtMn/9sCdnE0OJnrMMcnXPJy', null),
('Thủ kho', date 'now()' - interval '19 years', 'tk', '0000000003', '$2a$12$sPDZCZEc01jKDxKNDhZgquKZH.4R0TtMn/9sCdnE0OJnrMMcnXPJy', 1),
('Kế toán', date 'now()' - interval '19 years', 'kt', '0000000004', '$2a$12$sPDZCZEc01jKDxKNDhZgquKZH.4R0TtMn/9sCdnE0OJnrMMcnXPJy', 1),
('Nhân viên', date 'now()' - interval '19 years', 'nv', '0000000005', '$2a$12$sPDZCZEc01jKDxKNDhZgquKZH.4R0TtMn/9sCdnE0OJnrMMcnXPJy', 1),
('Nhân viên', date 'now()' - interval '19 years', 'nv', '0000000007', '$2a$12$sPDZCZEc01jKDxKNDhZgquKZH.4R0TtMn/9sCdnE0OJnrMMcnXPJy', 1)
;

update account set store_id = 1 where role = 'Admin';

insert into account (role, bdate, name, phone, password_hash, store_id) values
('Nhân viên', date 'now()' - interval '19 years', 'nv kho', '0000000006', '$2a$12$sPDZCZEc01jKDxKNDhZgquKZH.4R0TtMn/9sCdnE0OJnrMMcnXPJy', 1)
;

-- begin/ insert data from making a purchase to putting away the coresponding receives
insert into purchase (warehouse_id, account_id, supplier_id, expected_at, status) values
(1, 4, 1, timestamp 'now()' + interval '1 day', 'CHỜ PHẢN HỒI')
;

insert into purchase_item (purchase_id, gtin, quantity) values
(1, '619659115906', 5)
,(1, '8888021200126', 5)
;

insert into receive (purchase_id, account_id, expected_at, voucher_id) values
(1, 4, timestamp 'now()' + interval '1 day', 'VOU-001')
;

insert into receive_item (purchase_id, receive_id, gtin, quantity) values
(1, 1, '619659115906', 5)
,(1, 1, '8888021200126', 5)
;
update purchase set status = 'CHỜ NHẬP' where id = 1;

update receive set processed_by = 3, actual_at = timestamp 'now()' + interval '1 day' where id = 1;
insert into serial (nanoid, gtin, purchase_id, receive_id, receive_tote) values
('SER-ddEHD2fL3pynUGK4FZSUA', '8888021200126', 1, 1, 1)
,('SER-RDUuGi_UwzkYXls79aqCF', '8888021200126', 1, 1, 1)
,('SER-Ij2czMsg9ApYZI8kggcxL', '8888021200126', 1, 1, 1)
,('SER-G9jVbMaf5kMkjj_AbYwS-', '8888021200126', 1, 1, 1)
,('SER-XhjoVoESIan9Oy1NEm94v', '8888021200126', 1, 1, 1)
,('SER-rMiGVGZIj4uVYePjnSZwW', '619659115906', 1, 1, 1)
,('SER-0kzEXLSPez_NeSBybUT1A', '619659115906', 1, 1, 1)
,('SER-uBobwbyVeKLVFAOMw4tkA', '619659115906', 1, 1, 1)
,('SER-bR4nemcPjES_9RoR2OU5D', '619659115906', 1, 1, 1)
,('SER-QDVSSh0u3BeNh4fdq4Jtv', '619659115906', 1, 1, 1)
;
update purchase set status = 'KẾT THÚC' where id = 1;

update receive set putaway_by = 3, putaway_at = timestamp 'now()' + interval '1 day' where id = 1;
update serial set bin_id = 1
where nanoid in (
	'SER-ddEHD2fL3pynUGK4FZSUA'
	,'SER-RDUuGi_UwzkYXls79aqCF' 
	,'SER-Ij2czMsg9ApYZI8kggcxL'
	,'SER-G9jVbMaf5kMkjj_AbYwS-'
	,'SER-XhjoVoESIan9Oy1NEm94v'
	,'SER-rMiGVGZIj4uVYePjnSZwW'
	,'SER-0kzEXLSPez_NeSBybUT1A'
	,'SER-uBobwbyVeKLVFAOMw4tkA'
	,'SER-bR4nemcPjES_9RoR2OU5D'
	,'SER-QDVSSh0u3BeNh4fdq4Jtv'
);
-- end/ insert data from making a purchase to putting away the coresponding receives

-- begin/ insert OUT OF DATE data from making a purchase to putting away the coresponding receives
insert into purchase (warehouse_id, account_id, supplier_id, expected_at, status) values
(1, 4, 2, timestamp '2023-01-01' + interval '1 day', 'CHỜ PHẢN HỒI')
;

insert into purchase_item (purchase_id, gtin, quantity) values
(2, '4983435764166', 2)
,(2, '8936134272407', 2)
;

insert into receive (purchase_id, account_id, expected_at, voucher_id) values
(2, 4, timestamp '2023-01-01' + interval '1 day', 'VOU-001')
;

insert into receive_item (purchase_id, receive_id, gtin, quantity) values
(2, 2, '4983435764166', 2)
,(2, 2, '8936134272407', 2)
;
update purchase set status = 'CHỜ NHẬP' where id = 2;

update receive set processed_by = 3, actual_at = timestamp '2023-01-01' + interval '1 day' where id = 2;
insert into serial (nanoid, gtin, purchase_id, receive_id, receive_tote) values
('SER-NQW5vrGo6FFcO-iOl9Eu2', '4983435764166', 2, 2, 5)
,('SER-0qfVw995gkeV2bnUZdMJS', '4983435764166', 2, 2, 5)
,('SER-RHtfROpepYRFawn2AUVGp', '8936134272407', 2, 2, 5)
,('SER-IJWfZV9Kciem0mI5-sz-d', '8936134272407', 2, 2, 5)
;
update purchase set status = 'KẾT THÚC' where id = 2;

update receive set putaway_by = 3, putaway_at = timestamp '2023-01-01' + interval '2 day' where id = 2;
update serial set bin_id = 2
where nanoid in (
	'SER-NQW5vrGo6FFcO-iOl9Eu2'
	,'SER-0qfVw995gkeV2bnUZdMJS' 
	,'SER-RHtfROpepYRFawn2AUVGp'
	,'SER-IJWfZV9Kciem0mI5-sz-d'
);
-- end/ insert OUT OF DATE data from making a purchase to putting away the coresponding receives

-- begin/ add resupply ,export, pick export
insert into resupply (created_at, expected_at, account_id, store_id) values
(now(), timestamp 'now()' + interval '1 day', 7, 1)
;

insert into resupply_item (resupply_id, gtin, quantity) values
(1, 8888021200126, 1)
,(1, 619659115906, 1)
;

insert into export (account_id, created_at, expected_at, voucher_id, resupply_id) values
(7, now(), timestamp 'now()' + interval '1 day', 'VOU-1', 1)
;

insert into export_item (export_id, resupply_id, gtin, quantity) values
(1, 1, 8888021200126, 1)
,(1, 1, 619659115906, 1)
;

update resupply set status = 'CHỜ XUẤT' where id = 1;

update export set
picked_by = 3
,picked_at = now()
where id = 1
;

--SER-rMiGVGZIj4uVYePjnSZwW;619659115906
update serial set
pick_tote = 1
,export_id = 1
,resupply_id = 1
where nanoid in (
	'SER-G9jVbMaf5kMkjj_AbYwS-'
	,'SER-rMiGVGZIj4uVYePjnSZwW'
)
;

-- update export_item set pick_note = 'khong tim thay 1'
-- where export_id = 1
-- and resupply_id = 1
-- and gtin = '8888021200126'
-- ;

update resupply set status = 'KẾT THÚC' where id = 1;
-- end/ add resupply ,export, pick export


alter table warehouse add constraint pk_warehouse primary key(id);
alter table transfer add constraint pk_transfer primary key(id);
alter table inventory add constraint pk_inventory primary key(id);
alter table inventory_serial add constraint pk_inventory_serial primary key(inventory_id, serial);
alter table account add constraint pk_account primary key(id);
alter table store add constraint pk_store primary key(id);
alter table purchase add constraint pk_purchase primary key(id);
alter table purchase_item add constraint pk_purchase_item primary key(purchase_id, gtin);
alter table item add constraint pk_item primary key(gtin);
alter table resupply add constraint pk_resupply primary key(id);
alter table tote add constraint pk_tote primary key(id);
alter table bin add constraint pk_bin primary key(id);
alter table supplier add constraint pk_supplier primary key(id);
alter table supplier_item add constraint pk_supplier_item primary key(supplier_id, gtin);
alter table receive_item add constraint pk_receive_item primary key(receive_id, purchase_id, gtin);
alter table receive add constraint pk_receive primary key(id);
alter table export add constraint pk_export primary key(id);
alter table export_item add constraint pk_export_item primary key(export_id, resupply_id, gtin);
alter table resupply_item add constraint pk_resupply_item primary key(resupply_id, gtin);
alter table serial add constraint pk_serial primary key(nanoid);
alter table difference_serial add constraint pk_difference_serial primary key(nanoid);
alter table package add constraint pk_package primary key(nanoid);
alter table package_item add constraint pk_package_item primary key(package_nanoid, gtin);
alter table package_serial add constraint pk_package_serial primary key(package_nanoid, gtin, serial_nanoid);

alter table bin add constraint unq_bin_warehouse_id_shelf_row_col unique(warehouse_id, shelf, row, col);

alter table transfer add constraint fk_transfer_export_warehouse foreign key(export_warehouse) references warehouse(id);
alter table transfer add constraint fk_transfer_receive_warehouse foreign key(receive_warehouse) references warehouse(id);
alter table transfer add constraint fk_transfer_account_id foreign key(account_id) references account(id);

alter table inventory add constraint fk_inventory_warehouse_id foreign key(warehouse_id) references warehouse(id);
alter table inventory add constraint fk_inventory_account_id foreign key(account_id) references account(id);

alter table inventory_serial add constraint fk_inventory_serial_inventory foreign key(inventory_id) references inventory(id);
alter table inventory_serial add constraint fk_inventory_serial_serial foreign key(serial) references serial(nanoid);

alter table account add constraint fk_account_warehouse_id foreign key(warehouse_id) references warehouse(id);
alter table account add constraint fk_account_store_id foreign key(store_id) references store(id);

alter table store add constraint fk_store_warehouse_id foreign key(warehouse_id) references warehouse(id);

alter table tote add constraint fk_tote_warehouse_id foreign key(warehouse_id) references warehouse(id);
alter table bin add constraint fk_bin_warehouse_id foreign key(warehouse_id) references warehouse(id);

alter table purchase add constraint fk_purchase_warehouse_id foreign key(warehouse_id) references warehouse(id);
alter table purchase add constraint fk_purchase_supplier_id foreign key(supplier_id) references supplier(id);
alter table purchase add constraint fk_purchase_account_id foreign key(account_id) references account(id);
alter table purchase add constraint fk_purchase_receive_add_owner foreign key(receive_add_owner) references account(id);

alter table purchase_item add constraint fk_purchase_item_purchase_id foreign key(purchase_id) references purchase(id);
alter table purchase_item add constraint fk_purchase_item_gtin foreign key(gtin) references item(gtin);

alter table resupply add constraint fk_resupply_account_id foreign key(account_id) references account(id);
alter table resupply add constraint fk_resupply_export_add_owner foreign key(export_add_owner) references account(id);
alter table resupply add constraint fk_resupply_store_id foreign key(store_id) references store(id);

alter table supplier_item add constraint fk_supplier_gtin foreign key(gtin) references item(gtin);
alter table supplier_item add constraint fk_supplier_supplier_id foreign key(supplier_id) references supplier(id);

alter table receive_item add constraint fk_receive_item_receive_id foreign key(receive_id) references receive(id);
alter table receive_item add constraint fk_receive_item_purchase_id_gtin foreign key(purchase_id, gtin) references purchase_item(purchase_id, gtin);

alter table receive add constraint fk_receive_purchase_id foreign key(purchase_id) references purchase(id);
alter table receive add constraint fk_receive_account_id foreign key(account_id) references account(id);
alter table receive add constraint fk_receive_transfer_id foreign key(transfer_id) references transfer(id);
alter table receive add constraint fk_receive_processed_by foreign key(processed_by) references account(id);
alter table receive add constraint fk_receive_putaway_by foreign key(putaway_by) references account(id);

alter table export add constraint fk_export_resupply_id foreign key(resupply_id) references resupply(id);
alter table export add constraint fk_export_transfer_id foreign key(transfer_id) references transfer(id);
alter table export add constraint fk_receive_picked_by foreign key(picked_by) references account(id);
alter table export add constraint fk_receive_packed_by foreign key(packed_by) references account(id);

alter table export_item add constraint fk_export_item_export_id foreign key(export_id) references export(id);
alter table export_item add constraint fk_export_item_resupply_id_gtin foreign key(resupply_id, gtin) references resupply_item(resupply_id, gtin);

alter table resupply_item add constraint fk_resupply_item_resupply_id foreign key(resupply_id) references resupply(id);
alter table resupply_item add constraint fk_resupply_item_gtin foreign key(gtin) references item(gtin);

alter table serial add constraint fk_serial_receive_tote foreign key(receive_tote) references tote(id);
alter table serial add constraint fk_serial_pick_tote foreign key(pick_tote) references tote(id);
alter table serial add constraint fk_serial_bin_id foreign key(bin_id) references bin(id);
alter table serial add constraint fk_serial_receive_id_purchase_id_gtin foreign key(receive_id, purchase_id, gtin) references receive_item(receive_id, purchase_id, gtin);
alter table serial add constraint fk_serial_export_id_resupply_id_gtin foreign key(export_id, resupply_id, gtin) references export_item(export_id, resupply_id, gtin);

alter table difference_serial add constraint fk_difference_serial_receive_tote foreign key(receive_tote) references tote(id);
alter table difference_serial add constraint fk_difference_serial_pick_tote foreign key(pick_tote) references tote(id);
alter table difference_serial add constraint fk_difference_serial_bin_id foreign key(bin_id) references bin(id);
alter table difference_serial add constraint fk_difference_serial_receive_id_purchase_id_gtin foreign key(receive_id, purchase_id, gtin) references receive_item(receive_id, purchase_id, gtin);
alter table difference_serial add constraint fk_difference_serial_export_id_resupply_id_gtin foreign key(export_id, resupply_id, gtin) references export_item(export_id, resupply_id, gtin);

alter table package add constraint fk_package_export_id foreign key(export_id) references export(id);
alter table package_item add constraint fk_package_item_package_nanoid foreign key(package_nanoid) references package(nanoid);
alter table package_item add constraint fk_package_item_gtin foreign key(gtin) references item(gtin);
alter table package_serial add constraint fk_package_serial_package_nanoid_gtin foreign key(package_nanoid, gtin) references package_item(package_nanoid, gtin);
alter table package_serial add constraint fk_package_serial_serial_nanoid foreign key(serial_nanoid) references serial(nanoid);

CREATE TABLE sessions (
	token TEXT PRIMARY KEY,
	data BYTEA NOT NULL,
	expiry TIMESTAMPTZ NOT NULL
);

CREATE INDEX sessions_expiry_idx ON sessions (expiry);

commit;
