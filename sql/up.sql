begin;

create table account (
	id bigserial not null,
	role text not null,
	bdate date not null check(date_part('year', now()) - date_part('year', bdate) >= 18),
	name text not null,
	phone text not null unique check(length(phone) >= 10),
	password_hash bytea not null,
	-- warehouse_id bigint not null default 0,
	-- store_id bigint not null default 0,

	primary key(id)
);

create table warehouse (
	id bigserial not null,

	primary key(id)
);

-- alter table account add constraint fk_warehouse foreign key(warehouse_id) references warehouse(id);

insert into account (role, bdate, name, phone, password_hash) values
('Admin', date 'now()' - interval '19 years', 'tim', '0000000001', '$2a$12$sPDZCZEc01jKDxKNDhZgquKZH.4R0TtMn/9sCdnE0OJnrMMcnXPJy'),
('Thủ kho', date 'now()' - interval '19 years', 'tk', '0000000002', '$2a$12$sPDZCZEc01jKDxKNDhZgquKZH.4R0TtMn/9sCdnE0OJnrMMcnXPJy'),
('Kế toán trưởng', date 'now()' - interval '19 years', 'ktt', '0000000003', '$2a$12$sPDZCZEc01jKDxKNDhZgquKZH.4R0TtMn/9sCdnE0OJnrMMcnXPJy'),
('Kế toán', date 'now()' - interval '19 years', 'kt', '0000000004', '$2a$12$sPDZCZEc01jKDxKNDhZgquKZH.4R0TtMn/9sCdnE0OJnrMMcnXPJy'),
('Nhân viên', date 'now()' - interval '19 years', 'nv', '0000000005', '$2a$12$sPDZCZEc01jKDxKNDhZgquKZH.4R0TtMn/9sCdnE0OJnrMMcnXPJy')
;

commit;
