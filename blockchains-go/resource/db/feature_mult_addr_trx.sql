alter table fc_order
    add tx_type int default 1 not null comment '出账类型；1：单地址出账；2：多地址出账';

alter table fc_order
    add total_amount decimal(60, 18) default 0 not null comment '订单总金额，多地址出账使用';

alter table fc_order_hot
    add total_amount decimal(60, 18) default 0 not null comment '订单总金额，多地址出账使用';

alter table fc_order_hot
    add tx_type int default 1 not null comment '出账类型；1：单地址出账；2：多地址出账';


alter table fc_transfers_apply
    add tx_type int default 1 not null comment '出账类型；1：单地址出账；2：多地址出账';



create table fc_order_txs
(
    id             bigint comment 'id',
    seq_no         varchar(64)              not null comment '交易流水号，全局唯一',
    parent_seq_no  varchar(64)              not null comment '父流水号',
    tx_id          varchar(256) default ''  not null comment '交易哈希，未出哈希时为空',
    outer_order_no varchar(128) default ''  not null comment '外部订单号',
    inner_order_no varchar(128) default ''  not null comment '内部订单号',
    mch            varchar(36)  default ''  not null comment '商户代号',
    chain          varchar(24)  default ''  not null comment '链',
    coin_code      varchar(24)  default ''  not null comment '代币编码',
    contract       varchar(256) default ''  not null comment '合约地址',
    from_address   varchar(256) default ''  not null comment 'from地址',
    to_address     varchar(256) default ''  not null comment 'to地址',
    amount         varchar(36)  default '0' not null comment '交易金额',
    status         int          default -1  not null comment '交易状态',
    sort           int                      not null comment '排序，降序排列',
    signer_no      varchar(24)  default ''  not null comment '签名机编号',
    create_time    datetime                 not null comment '创建时间',
    update_time    datetime                 not null comment '最后更新时间'
) comment '订单相关交易';

create
unique index fc_order_txs_id_uindex
	on fc_order_txs (id);

create
index fc_order_txs_seq_no_index
	on fc_order_txs (seq_no);

create
index fc_order_txs_tx_id_index
	on fc_order_txs (tx_id);

alter table fc_order_txs
    add constraint fc_order_txs_pk
        primary key (id);

alter table fc_order_txs modify id bigint auto_increment comment 'id';

alter table fc_order_txs
    add constraint fc_order_txs_pk
        primary key (id);


alter table fc_order_txs
    add sign_req_data longtext null comment '待签名数据' after signer_no;


alter table fc_order_txs
    add err_msg varchar(2048) default '' not null comment '错误信息' after sign_req_data;


alter table fc_order_txs
    add nonce bigint default -1 not null comment '本次交易使用的随机数' after sort;

alter table fc_order_txs
    add freeze_unlock int default 0 not null comment '冻结的金额是否已解锁，0：未解锁；1：已解锁' after err_msg;


drop
index fc_order_txs_seq_no_index on fc_order_txs;

create
unique index fc_order_txs_seq_no_uindex
	on fc_order_txs (seq_no);

alter table fc_order_txs modify amount decimal (60,18) default 0 not null comment '交易金额';



create table fc_order_txs_push
(
    id           bigint primary key auto_increment comment 'id',
    order_txs_id bigint comment '订单交易id',
    tx_id        varchar(256) default '' not null comment '交易哈希，未出哈希时为空',
    block_height bigint       default 0  not null comment '当前区块高度',
    is_in        int                     not null comment '是否入账；1：入账，2：出账',
    confirmation int          default 0  not null comment '确认数',
    confirm_time bigint                  not null comment '确认时间',
    fee          varchar(36)             not null comment '手续费',
    memo         varchar(512)            not null comment '交易备注',
    trx_n        int          default -1 not null comment '交易下标',
    create_time  datetime                not null comment '创建时间'
) comment '订单相关交易推送';


create
unique index fc_order_txs_push_order_txs_id_uindex
	on fc_order_txs_push (order_txs_id);


alter table fc_transfers_apply_coin_address
    add ban_from_address varchar(255) default '' not null comment '不可使用此地址出账';

