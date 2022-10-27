create table fc_order_priority
(
    id int,
    apply_id int not null comment '对应fc_transfers_apply表的 id',
    outer_order_no varchar(64) not null comment '订单外部编号',
    chain_name varchar(16) not null comment '链名',
    create_time datetime not null comment '添加时间'
);

create unique index fc_order_priority_apply_id_uindex
    on fc_order_priority (apply_id);

create unique index fc_order_priority_id_uindex
    on fc_order_priority (id);

alter table fc_order_priority
    add constraint fc_order_priority_pk
        primary key (id);

alter table fc_order_priority modify id int auto_increment;

alter table fc_order_priority
    add status int default 1 not null comment '1:正在处理，2:订单已完成' after chain_name;

create index fc_order_priority_outer_order_no_index
    on fc_order_priority (outer_order_no);


alter table fc_order_priority
    add coin_code varchar(24) default '' not null after chain_name;

alter table fc_order_priority
    add mch_id int default 0 not null after coin_code;

