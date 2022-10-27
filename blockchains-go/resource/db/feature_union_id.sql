alter table fc_coin_set
    add union_id int default 0 not null comment '全局id，与交易所统一';

