-- MySQL dump 10.13  Distrib 5.7.31, for Linux (x86_64)
--
-- Host: 172.17.2.65    Database: zecsync
-- ------------------------------------------------------
-- Server version	5.7.30-0ubuntu0.18.04.1

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `notifyresult`
--

DROP TABLE IF EXISTS `notifyresult`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `notifyresult` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `userid` int(11) NOT NULL DEFAULT '0' COMMENT 'é€šçŸ¥ç”¨æˆ·id',
  `height` bigint(20) NOT NULL COMMENT 'é«˜åº¦',
  `txid` varchar(255) NOT NULL DEFAULT '' COMMENT 'äº¤æ˜“id',
  `num` int(11) NOT NULL DEFAULT '0' COMMENT 'æŽ¨é€æ¬¡æ•°',
  `timestamp` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'æŽ¨é€æ—¶é—´',
  `result` int(11) NOT NULL DEFAULT '0' COMMENT 'æŽ¨é€ç»“æžœ 1è¡¨ç¤ºæˆåŠŸ',
  `content` varchar(1024) NOT NULL DEFAULT '' COMMENT 'å¤±è´¥å†…å®¹',
  `type` tinyint(3) DEFAULT NULL COMMENT '0ä¸ºæ™®é€šäº¤æ˜“æŽ¨é€ï¼Œï¼‘ä¸ºç¡®è®¤æ•°æŽ¨é€',
  PRIMARY KEY (`id`),
  KEY `userid` (`userid`,`txid`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COMMENT='事件通知表';
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2020-08-05 10:59:32
