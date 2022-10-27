-- Create syntax for TABLE 'block_info'
CREATE TABLE `block_info` (
                              `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
                              `height` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '区块高度',
                              `hash` varchar(100) NOT NULL DEFAULT '' COMMENT '区块hash值',
                              `previousblockhash` varchar(100) NOT NULL DEFAULT '' COMMENT '前一个区块hash',
                              `nextblockhash` varchar(100) DEFAULT NULL COMMENT '后一个区块hash',
                              `timestamp` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '时间戳',
                              `transactions` int(11) NOT NULL DEFAULT '0' COMMENT '交易总数',
                              `confirmations` int(11) DEFAULT NULL COMMENT '确认数',
                              `createtime` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00' COMMENT '记录时间',
                              PRIMARY KEY (`id`) USING BTREE,
                              UNIQUE KEY `idx_hash` (`hash`) USING BTREE,
                              KEY `idx_height` (`height`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 ROW_FORMAT=DYNAMIC COMMENT='块信息表';

-- Create syntax for TABLE 'block_tx'
CREATE TABLE `block_tx` (
                            `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
                            `coin_name` varchar(20) NOT NULL COMMENT '币种名称',
                            `txid` varchar(100) NOT NULL DEFAULT '' COMMENT '交易hash',
                            `contract_address` varchar(100) NOT NULL DEFAULT '' COMMENT '合约地址',
                            `from_address` varchar(100) NOT NULL DEFAULT '' COMMENT '转出地址',
                            `to_address` varchar(100) NOT NULL DEFAULT '' COMMENT '接收地址',
                            `block_height` bigint(20) NOT NULL COMMENT '高度',
                            `block_hash` varchar(100) NOT NULL COMMENT '块hash',
                            `amount` decimal(50,0) NOT NULL COMMENT '金额',
                            `status` tinyint(3) DEFAULT NULL COMMENT '0代表 失败,1代表成功,2代表上链成功但交易失败',
                            `gas_used` bigint(20) NOT NULL,
                            `gas_price` bigint(20) NOT NULL,
                            `nonce` bigint(20) NOT NULL,
                            `input` text,
                            `logs` text,
                            `decimal` tinyint(3) DEFAULT NULL COMMENT '币种精度',
                            `timestamp` timestamp NULL DEFAULT NULL COMMENT '交易时间戳',
                            `create_time` timestamp NULL DEFAULT NULL COMMENT '创建时间',
                            `fee` decimal(50,0) NOT NULL COMMENT 'fee',
                            `memo` varchar(100) NOT NULL DEFAULT '' COMMENT 'memo',
                            PRIMARY KEY (`id`),
                            KEY `contract` (`contract_address`) USING BTREE,
                            KEY `from` (`from_address`) USING BTREE,
                            KEY `to` (`to_address`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;
-- Create syntax for TABLE 'notifyresult'
CREATE TABLE `notifyresult` (
                                `id` bigint(20) NOT NULL AUTO_INCREMENT,
                                `userid` int(11) NOT NULL DEFAULT '0' COMMENT '通知用户id',
                                `height` bigint(20) NOT NULL COMMENT '高度',
                                `txid` varchar(255) NOT NULL DEFAULT '' COMMENT '交易id',
                                `num` int(11) NOT NULL DEFAULT '0' COMMENT '推送次数',
                                `timestamp` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '推送时间',
                                `result` int(11) NOT NULL DEFAULT '0' COMMENT '推送结果 1表示成功',
                                `content` varchar(1024) NOT NULL DEFAULT '' COMMENT '失败内容',
                                `type` tinyint(3) DEFAULT NULL COMMENT '0为普通交易推送，１为确认数推送',
                                PRIMARY KEY (`id`),
                                KEY `userid` (`userid`,`txid`(191))
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT='事件通知表';