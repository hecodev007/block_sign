CREATE TABLE `block_info` (
                              `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
                              `height` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '区块高度',
                              `hash` varchar(100) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '区块hash值',
                              `version` int(11) NOT NULL DEFAULT '0' COMMENT '版本',
                              `previousblockhash` varchar(100) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
                              `nextblockhash` varchar(100) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
                              `timestamp` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '时间戳',
                              `transactions` int(11) NOT NULL DEFAULT '0' COMMENT '交易总数',
                              `confirmations` int(11) NOT NULL COMMENT '确认数',
                              `createtime` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                              `forked` tinyint(4) DEFAULT NULL COMMENT '是否分叉',
                              PRIMARY KEY (`id`),
                              UNIQUE KEY `idx_hash` (`hash`),
                              KEY `idx_height` (`height`)
) ENGINE=InnoDB AUTO_INCREMENT=121827 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='块信息表'

CREATE TABLE `block_tx` (
                            `id` bigint(20) NOT NULL AUTO_INCREMENT,
                            `txid` varchar(100) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '交易id',
                            `height` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '区块高度索引值',
                            `blockhash` varchar(100) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '区块hash值',
                            `version` int(11) NOT NULL DEFAULT '0',
                            `fee` decimal(40,8) NOT NULL DEFAULT '0.00000000',
                            `vincount` int(11) NOT NULL DEFAULT '0',
                            `voutcount` int(11) NOT NULL DEFAULT '0',
                            `timestamp` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '交易时间戳',
                            `createtime` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                            `forked` tinyint(4) DEFAULT NULL COMMENT '是否因为块分叉无效',
                            `iscoinbase` tinyint(4) DEFAULT NULL COMMENT '是否因为创世块',
                            PRIMARY KEY (`id`),
                            UNIQUE KEY `idx_txid` (`txid`),
                            KEY `idx_highindex` (`height`)
) ENGINE=InnoDB AUTO_INCREMENT=44 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='交易表'


CREATE TABLE `block_tx_vin` (
                                `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
                                `message_id` varchar(100) COLLATE utf8mb4_bin NOT NULL,
                                `txid` varchar(100) COLLATE utf8mb4_bin NOT NULL,
                                `output_index` bigint(20) unsigned NOT NULL DEFAULT '0',
                                `ledger_index` bigint(20) unsigned NOT NULL DEFAULT '0',
                                `spent` tinyint(4) DEFAULT NULL,
                                `value` varchar(64) COLLATE utf8mb4_bin NOT NULL,
                                `address` varchar(100) COLLATE utf8mb4_bin NOT NULL,
                                `createtime` timestamp NULL DEFAULT NULL,
                                `height` bigint(20) NOT NULL DEFAULT '0',
                                `forked` int(11) DEFAULT NULL COMMENT '是否分叉成孤块',
                                PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=48 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin



CREATE TABLE `block_tx_vout` (
                                 `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
                                 `message_id` varchar(100) COLLATE utf8mb4_bin NOT NULL,
                                 `txid` varchar(100) COLLATE utf8mb4_bin NOT NULL,
                                 `output_index` bigint(20) unsigned NOT NULL DEFAULT '0',
                                 `ledger_index` bigint(20) unsigned NOT NULL DEFAULT '0',
                                 `spent` tinyint(4) DEFAULT NULL,
                                 `value` varchar(64) COLLATE utf8mb4_bin NOT NULL,
                                 `address` varchar(100) COLLATE utf8mb4_bin NOT NULL,
                                 `createtime` timestamp NULL DEFAULT NULL,
                                 `height` bigint(20) NOT NULL DEFAULT '0',
                                 `forked` int(11) DEFAULT NULL COMMENT '是否分叉成孤块',
                                 PRIMARY KEY (`id`),
                                 KEY `addr` (`address`),
                                 KEY `txid` (`txid`)
) ENGINE=InnoDB AUTO_INCREMENT=66 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='交易输出表'


CREATE TABLE `notifyresult` (
                                `id` bigint(20) NOT NULL AUTO_INCREMENT,
                                `userid` int(11) NOT NULL DEFAULT '0' COMMENT '通知用户id',
                                `txid` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '交易id',
                                `num` int(11) NOT NULL DEFAULT '0' COMMENT '推送次数',
                                `timestamp` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '推送时间',
                                `result` int(11) NOT NULL DEFAULT '0' COMMENT '推送结果 1表示成功',
                                `content` varchar(1024) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '失败内容',
                                `height` bigint(20) DEFAULT NULL,
                                `type` int(11) DEFAULT NULL,
                                PRIMARY KEY (`id`),
                                KEY `userid` (`userid`,`txid`(191))
) ENGINE=InnoDB AUTO_INCREMENT=242 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='事件通知表'

