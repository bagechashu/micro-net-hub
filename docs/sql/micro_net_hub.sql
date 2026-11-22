-- MySQL dump 10.13  Distrib 8.0.46, for Linux (x86_64)
--
-- Host: localhost    Database: micro_net_hub
-- ------------------------------------------------------
-- Server version	8.0.46

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!50503 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `apis`
--

DROP TABLE IF EXISTS `apis`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `apis` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `method` varchar(20) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '''请求方式''',
  `path` varchar(100) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '''访问路径''',
  `category` varchar(50) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '''所属类别''',
  `remark` varchar(100) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '''备注''',
  `creator` varchar(20) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '''创建人''',
  PRIMARY KEY (`id`),
  KEY `idx_apis_deleted_at` (`deleted_at`)
) ENGINE=InnoDB AUTO_INCREMENT=75 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `apis`
--

LOCK TABLES `apis` WRITE;
/*!40000 ALTER TABLE `apis` DISABLE KEYS */;
INSERT INTO `apis` VALUES (1,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/base/login','base','用户登录','System'),(2,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/base/logout','base','用户登出','System'),(3,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'GET','/base/refreshToken','base','刷新JWT令牌','System'),(4,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/base/sendcode','base','给用户邮箱发送验证码','System'),(5,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/base/changePwd','base','通过邮箱验证码重置密码','System'),(6,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'GET','/user/info','user','获取当前登录用户信息','System'),(7,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'GET','/user/list','user','获取用户列表','System'),(8,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/user/resetTotpSecret','user','重置用户 TOTP 秘钥','System'),(9,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/user/changePwd','user','更新用户登录密码','System'),(10,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/user/add','user','创建用户','System'),(11,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/user/update','user','更新用户','System'),(12,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/user/delete','user','批量删除用户','System'),(13,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/user/changeUserStatus','user','更改用户在职状态','System'),(14,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/goldap/sync/syncDingTalkUsers','user','从钉钉拉取用户信息','System'),(15,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/goldap/sync/syncWeComUsers','user','从企业微信拉取用户信息','System'),(16,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/goldap/sync/syncFeiShuUsers','user','从飞书拉取用户信息','System'),(17,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/goldap/sync/syncOpenLdapUsers','user','从openldap拉取用户信息','System'),(18,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/goldap/sync/syncSqlUsers','user','将数据库中的用户同步到Ldap','System'),(19,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'GET','/group/list','group','获取分组列表','System'),(20,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'GET','/group/tree','group','获取分组列表树','System'),(21,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/group/add','group','创建分组','System'),(22,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/group/update','group','更新分组','System'),(23,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/group/delete','group','批量删除分组','System'),(24,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/group/adduser','group','添加用户到分组','System'),(25,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/group/removeuser','group','将用户从分组移出','System'),(26,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'GET','/group/useringroup','group','获取在分组内的用户列表','System'),(27,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'GET','/group/usernoingroup','group','获取不在分组内的用户列表','System'),(28,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/goldap/sync/syncDingTalkDepts','group','从钉钉拉取部门信息','System'),(29,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/goldap/sync/syncWeComDepts','group','从企业微信拉取部门信息','System'),(30,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/goldap/sync/syncFeiShuDepts','group','从飞书拉取部门信息','System'),(31,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/goldap/sync/syncOpenLdapDepts','group','从openldap拉取部门信息','System'),(32,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/goldap/sync/syncSqlGroups','group','将数据库中的分组同步到Ldap','System'),(33,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'GET','/role/list','role','获取角色列表','System'),(34,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/role/add','role','创建角色','System'),(35,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/role/update','role','更新角色','System'),(36,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'GET','/role/getmenulist','role','获取角色的权限菜单','System'),(37,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/role/updatemenus','role','更新角色的权限菜单','System'),(38,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'GET','/role/getapilist','role','获取角色的权限接口','System'),(39,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/role/updateapis','role','更新角色的权限接口','System'),(40,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/role/delete','role','批量删除角色','System'),(41,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'GET','/menu/tree','menu','获取菜单树','System'),(42,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'GET','/menu/access/tree','menu','获取用户菜单树','System'),(43,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/menu/add','menu','创建菜单','System'),(44,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/menu/update','menu','更新菜单','System'),(45,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/menu/delete','menu','批量删除菜单','System'),(46,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'GET','/api/list','api','获取接口列表','System'),(47,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'GET','/api/tree','api','获取接口树','System'),(48,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/api/add','api','创建接口','System'),(49,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/api/update','api','更新接口','System'),(50,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/api/delete','api','批量删除接口','System'),(51,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'GET','/goldap/fieldrelation/list','fieldrelation','获取字段动态关系列表','System'),(52,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/goldap/fieldrelation/add','fieldrelation','创建字段动态关系','System'),(53,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/goldap/fieldrelation/update','fieldrelation','更新字段动态关系','System'),(54,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/goldap/fieldrelation/delete','fieldrelation','批量删除字段动态关系','System'),(55,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'GET','/log/operation/list','log','获取操作日志列表','System'),(56,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/log/operation/delete','log','批量删除操作日志','System'),(57,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'GET','/sitenav/getnav','sitenav','导航页-获取全部分组和站点','System'),(58,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/sitenav/group/add','sitenav','导航页-新增分组','System'),(59,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/sitenav/group/update','sitenav','导航页-更新分组','System'),(60,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/sitenav/group/delete','sitenav','导航页-删除分组','System'),(61,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/sitenav/site/add','sitenav','导航页-新增站点','System'),(62,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/sitenav/site/update','sitenav','导航页-更新站点','System'),(63,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/sitenav/site/delete','sitenav','导航页-删除站点','System'),(64,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'GET','/dns/getall','dns','DNS-获取全部域名及解析记录','System'),(65,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/dns/zone/add','dns','DNS-新增域名(Zone)','System'),(66,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/dns/zone/update','dns','DNS-更新域名(Zone)','System'),(67,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/dns/zone/delete','dns','DNS-删除域名(Zone)','System'),(68,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/dns/record/add','dns','DNS-新增解析记录','System'),(69,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/dns/record/update','dns','DNS-更新解析记录','System'),(70,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/dns/record/delete','dns','DNS-删除解析记录','System'),(71,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'GET','/notice/getall','noticeboard','首页公告-获取所有公告','System'),(72,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/notice/mgr/add','noticeboard','首页公告-增加公告','System'),(73,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/notice/mgr/update','noticeboard','首页公告-更新公告','System'),(74,'2026-11-22 15:03:16.195','2026-11-22 15:03:16.195',NULL,'POST','/notice/mgr/delete','noticeboard','首页公告-删除公告','System');
/*!40000 ALTER TABLE `apis` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `casbin_rule`
--

DROP TABLE IF EXISTS `casbin_rule`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `casbin_rule` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `ptype` varchar(100) COLLATE utf8mb4_general_ci DEFAULT NULL,
  `v0` varchar(100) COLLATE utf8mb4_general_ci DEFAULT NULL,
  `v1` varchar(100) COLLATE utf8mb4_general_ci DEFAULT NULL,
  `v2` varchar(100) COLLATE utf8mb4_general_ci DEFAULT NULL,
  `v3` varchar(100) COLLATE utf8mb4_general_ci DEFAULT NULL,
  `v4` varchar(100) COLLATE utf8mb4_general_ci DEFAULT NULL,
  `v5` varchar(100) COLLATE utf8mb4_general_ci DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_casbin_rule` (`ptype`,`v0`,`v1`,`v2`,`v3`,`v4`,`v5`)
) ENGINE=InnoDB AUTO_INCREMENT=86 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `casbin_rule`
--

LOCK TABLES `casbin_rule` WRITE;
/*!40000 ALTER TABLE `casbin_rule` DISABLE KEYS */;
INSERT INTO `casbin_rule` VALUES (57,'p','superadmin','/api/add','POST','','',''),(59,'p','superadmin','/api/delete','POST','','',''),(55,'p','superadmin','/api/list','GET','','',''),(56,'p','superadmin','/api/tree','GET','','',''),(58,'p','superadmin','/api/update','POST','','',''),(9,'p','superadmin','/base/changePwd','POST','','',''),(1,'p','superadmin','/base/login','POST','','',''),(3,'p','superadmin','/base/logout','POST','','',''),(5,'p','superadmin','/base/refreshToken','GET','','',''),(7,'p','superadmin','/base/sendcode','POST','','',''),(74,'p','superadmin','/dns/getall','GET','','',''),(78,'p','superadmin','/dns/record/add','POST','','',''),(80,'p','superadmin','/dns/record/delete','POST','','',''),(79,'p','superadmin','/dns/record/update','POST','','',''),(75,'p','superadmin','/dns/zone/add','POST','','',''),(77,'p','superadmin','/dns/zone/delete','POST','','',''),(76,'p','superadmin','/dns/zone/update','POST','','',''),(61,'p','superadmin','/goldap/fieldrelation/add','POST','','',''),(63,'p','superadmin','/goldap/fieldrelation/delete','POST','','',''),(60,'p','superadmin','/goldap/fieldrelation/list','GET','','',''),(62,'p','superadmin','/goldap/fieldrelation/update','POST','','',''),(36,'p','superadmin','/goldap/sync/syncDingTalkDepts','POST','','',''),(22,'p','superadmin','/goldap/sync/syncDingTalkUsers','POST','','',''),(38,'p','superadmin','/goldap/sync/syncFeiShuDepts','POST','','',''),(24,'p','superadmin','/goldap/sync/syncFeiShuUsers','POST','','',''),(39,'p','superadmin','/goldap/sync/syncOpenLdapDepts','POST','','',''),(25,'p','superadmin','/goldap/sync/syncOpenLdapUsers','POST','','',''),(40,'p','superadmin','/goldap/sync/syncSqlGroups','POST','','',''),(26,'p','superadmin','/goldap/sync/syncSqlUsers','POST','','',''),(37,'p','superadmin','/goldap/sync/syncWeComDepts','POST','','',''),(23,'p','superadmin','/goldap/sync/syncWeComUsers','POST','','',''),(29,'p','superadmin','/group/add','POST','','',''),(32,'p','superadmin','/group/adduser','POST','','',''),(31,'p','superadmin','/group/delete','POST','','',''),(27,'p','superadmin','/group/list','GET','','',''),(33,'p','superadmin','/group/removeuser','POST','','',''),(28,'p','superadmin','/group/tree','GET','','',''),(30,'p','superadmin','/group/update','POST','','',''),(34,'p','superadmin','/group/useringroup','GET','','',''),(35,'p','superadmin','/group/usernoingroup','GET','','',''),(65,'p','superadmin','/log/operation/delete','POST','','',''),(64,'p','superadmin','/log/operation/list','GET','','',''),(50,'p','superadmin','/menu/access/tree','GET','','',''),(52,'p','superadmin','/menu/add','POST','','',''),(54,'p','superadmin','/menu/delete','POST','','',''),(49,'p','superadmin','/menu/tree','GET','','',''),(53,'p','superadmin','/menu/update','POST','','',''),(81,'p','superadmin','/notice/getall','GET','','',''),(83,'p','superadmin','/notice/mgr/add','POST','','',''),(85,'p','superadmin','/notice/mgr/delete','POST','','',''),(84,'p','superadmin','/notice/mgr/update','POST','','',''),(42,'p','superadmin','/role/add','POST','','',''),(48,'p','superadmin','/role/delete','POST','','',''),(46,'p','superadmin','/role/getapilist','GET','','',''),(44,'p','superadmin','/role/getmenulist','GET','','',''),(41,'p','superadmin','/role/list','GET','','',''),(43,'p','superadmin','/role/update','POST','','',''),(47,'p','superadmin','/role/updateapis','POST','','',''),(45,'p','superadmin','/role/updatemenus','POST','','',''),(66,'p','superadmin','/sitenav/getnav','GET','','',''),(68,'p','superadmin','/sitenav/group/add','POST','','',''),(70,'p','superadmin','/sitenav/group/delete','POST','','',''),(69,'p','superadmin','/sitenav/group/update','POST','','',''),(71,'p','superadmin','/sitenav/site/add','POST','','',''),(73,'p','superadmin','/sitenav/site/delete','POST','','',''),(72,'p','superadmin','/sitenav/site/update','POST','','',''),(18,'p','superadmin','/user/add','POST','','',''),(16,'p','superadmin','/user/changePwd','POST','','',''),(21,'p','superadmin','/user/changeUserStatus','POST','','',''),(20,'p','superadmin','/user/delete','POST','','',''),(11,'p','superadmin','/user/info','GET','','',''),(13,'p','superadmin','/user/list','GET','','',''),(14,'p','superadmin','/user/resetTotpSecret','POST','','',''),(19,'p','superadmin','/user/update','POST','','',''),(10,'p','user','/base/changePwd','POST','','',''),(2,'p','user','/base/login','POST','','',''),(4,'p','user','/base/logout','POST','','',''),(6,'p','user','/base/refreshToken','GET','','',''),(8,'p','user','/base/sendcode','POST','','',''),(51,'p','user','/menu/access/tree','GET','','',''),(82,'p','user','/notice/getall','GET','','',''),(67,'p','user','/sitenav/getnav','GET','','',''),(17,'p','user','/user/changePwd','POST','','',''),(12,'p','user','/user/info','GET','','',''),(15,'p','user','/user/resetTotpSecret','POST','','','');
/*!40000 ALTER TABLE `casbin_rule` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `dns_records`
--

DROP TABLE IF EXISTS `dns_records`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `dns_records` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `zone_id` bigint unsigned NOT NULL,
  `type` varchar(64) COLLATE utf8mb4_general_ci NOT NULL,
  `host` varchar(64) COLLATE utf8mb4_general_ci NOT NULL,
  `value` varchar(64) COLLATE utf8mb4_general_ci NOT NULL,
  `ttl` int unsigned NOT NULL,
  `creator` varchar(20) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '''创建人''',
  PRIMARY KEY (`id`),
  KEY `idx_dns_records_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `dns_records`
--

LOCK TABLES `dns_records` WRITE;
/*!40000 ALTER TABLE `dns_records` DISABLE KEYS */;
/*!40000 ALTER TABLE `dns_records` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `dns_zones`
--

DROP TABLE IF EXISTS `dns_zones`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `dns_zones` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `name` varchar(64) COLLATE utf8mb4_general_ci NOT NULL,
  `creator` varchar(20) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '''创建人''',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uni_dns_zones_name` (`name`),
  KEY `idx_dns_zones_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `dns_zones`
--

LOCK TABLES `dns_zones` WRITE;
/*!40000 ALTER TABLE `dns_zones` DISABLE KEYS */;
/*!40000 ALTER TABLE `dns_zones` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `field_relations`
--

DROP TABLE IF EXISTS `field_relations`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `field_relations` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `flag` longtext COLLATE utf8mb4_general_ci,
  `attributes` json DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_field_relations_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `field_relations`
--

LOCK TABLES `field_relations` WRITE;
/*!40000 ALTER TABLE `field_relations` DISABLE KEYS */;
/*!40000 ALTER TABLE `field_relations` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `group_users`
--

DROP TABLE IF EXISTS `group_users`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `group_users` (
  `group_id` bigint unsigned NOT NULL,
  `user_id` bigint unsigned NOT NULL,
  PRIMARY KEY (`group_id`,`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `group_users`
--

LOCK TABLES `group_users` WRITE;
/*!40000 ALTER TABLE `group_users` DISABLE KEYS */;
/*!40000 ALTER TABLE `group_users` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `groups`
--

DROP TABLE IF EXISTS `groups`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `groups` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `group_name` varchar(128) COLLATE utf8mb4_general_ci NOT NULL COMMENT '''分组名称''',
  `remark` varchar(128) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '''分组中文说明''',
  `creator` varchar(20) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '''创建人''',
  `parent_id` bigint unsigned DEFAULT '0' COMMENT '''父组编号(编号为0时表示根组)''',
  `group_type` varchar(20) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '''分组类型：cn、ou''',
  `source` varchar(20) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '''来源：dingTalk、weCom、ldap、platform''',
  `source_dept_id` varchar(100) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '''部门编号''',
  `source_dept_parent_id` varchar(100) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '''父部门编号''',
  `source_user_num` bigint DEFAULT '0' COMMENT '''部门下的用户数量，从第三方获取的数据''',
  `group_dn` varchar(255) COLLATE utf8mb4_general_ci NOT NULL COMMENT '''分组dn''',
  `sync_state` tinyint(1) DEFAULT '1' COMMENT '''同步状态:1已同步, 2未同步''',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uni_groups_group_name` (`group_name`),
  KEY `idx_groups_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `groups`
--

LOCK TABLES `groups` WRITE;
/*!40000 ALTER TABLE `groups` DISABLE KEYS */;
/*!40000 ALTER TABLE `groups` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `menus`
--

DROP TABLE IF EXISTS `menus`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `menus` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `name` varchar(50) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '''菜单名称(英文名, 可用于国际化)''',
  `icon` varchar(50) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '''菜单图标''',
  `path` varchar(100) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '''菜单访问路径''',
  `redirect` varchar(100) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '''重定向路径''',
  `component` varchar(100) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '''前端组件路径''',
  `sort` int DEFAULT '999' COMMENT '''菜单顺序(1-999)''',
  `status` tinyint(1) DEFAULT '1' COMMENT '''菜单状态(正常/禁用, 默认正常)''',
  `hidden` tinyint(1) DEFAULT '2' COMMENT '''菜单在侧边栏隐藏(1隐藏，2显示)''',
  `no_cache` tinyint(1) DEFAULT '2' COMMENT '''菜单是否被 <keep-alive> 缓存(1不缓存，2缓存)''',
  `always_show` tinyint(1) DEFAULT '2' COMMENT '''忽略之前定义的规则，一直显示根路由(1忽略，2不忽略)''',
  `breadcrumb` tinyint(1) DEFAULT '1' COMMENT '''面包屑可见性(可见/隐藏, 默认可见)''',
  `active_menu` varchar(100) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '''在其它路由时，想在侧边栏高亮的路由''',
  `parent_id` bigint unsigned DEFAULT '0' COMMENT '''父菜单编号(编号为0时表示根菜单)''',
  `creator` varchar(20) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '''创建人''',
  PRIMARY KEY (`id`),
  KEY `idx_menus_deleted_at` (`deleted_at`)
) ENGINE=InnoDB AUTO_INCREMENT=18 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `menus`
--

LOCK TABLES `menus` WRITE;
/*!40000 ALTER TABLE `menus` DISABLE KEYS */;
INSERT INTO `menus` VALUES (1,'2026-11-22 15:03:16.174','2026-11-22 15:03:16.174',NULL,'UserManage','user','/personnel','/personnel/user','Layout',1,1,2,2,2,1,'',0,'System'),(2,'2026-11-22 15:03:16.174','2026-11-22 15:03:16.174',NULL,'System','component','/system','/system/role','Layout',2,1,2,2,2,1,'',0,'System'),(3,'2026-11-22 15:03:16.174','2026-11-22 15:03:16.174',NULL,'Log','example','/log','/log/operation-log','Layout',3,1,2,2,2,1,'',0,'System'),(4,'2026-11-22 15:03:16.174','2026-11-22 15:03:16.174',NULL,'Profile','profile','/profile/index','','Layout',4,1,2,2,2,1,'',0,'System'),(5,'2026-11-22 15:03:16.174','2026-11-22 15:03:16.174',NULL,'User','people','user','','/personnel/user/index',11,1,2,1,2,1,'',1,'System'),(6,'2026-11-22 15:03:16.174','2026-11-22 15:03:16.174',NULL,'Group','peoples','group','','/personnel/group/index',12,1,2,1,2,1,'',1,'System'),(7,'2026-11-22 15:03:16.174','2026-11-22 15:03:16.174',NULL,'FieldRelation','relation','fieldRelation','','/personnel/fieldRelation/index',13,1,2,2,2,1,'',1,'System'),(8,'2026-11-22 15:03:16.174','2026-11-22 15:03:16.174',NULL,'Role','eye-open','role','','/system/role/index',21,1,2,2,2,1,'',2,'System'),(9,'2026-11-22 15:03:16.174','2026-11-22 15:03:16.174',NULL,'Menu','tree-table','menu','','/system/menu/index',22,1,2,2,2,1,'',2,'System'),(10,'2026-11-22 15:03:16.174','2026-11-22 15:03:16.174',NULL,'Api','tree','api','','/system/api/index',23,1,2,2,2,1,'',2,'System'),(11,'2026-11-22 15:03:16.174','2026-11-22 15:03:16.174',NULL,'Sitenav','list','sitenavmgr','','/sitenav/manager',24,1,2,2,2,1,'',2,'System'),(12,'2026-11-22 15:03:16.174','2026-11-22 15:03:16.174',NULL,'DnsManager','guide','dnsmgr','','/dns/manager',25,1,2,1,2,1,'',2,'System'),(13,'2026-11-22 15:03:16.174','2026-11-22 15:03:16.174',NULL,'NoticeManager','message','noticemgr','','/notice/manager',26,1,2,1,2,1,'',2,'System'),(14,'2026-11-22 15:03:16.174','2026-11-22 15:03:16.174',NULL,'OperationLog','documentation','operation-log','','/log/operation-log/index',31,1,2,2,2,1,'',3,'System');
/*!40000 ALTER TABLE `menus` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `nav_groups`
--

DROP TABLE IF EXISTS `nav_groups`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `nav_groups` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `title` varchar(20) COLLATE utf8mb4_general_ci NOT NULL COMMENT '''网址分组标题''',
  `name` varchar(20) COLLATE utf8mb4_general_ci NOT NULL COMMENT '''网址分组名''',
  `creator` varchar(20) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '''创建人''',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uni_nav_groups_name` (`name`),
  KEY `idx_nav_groups_deleted_at` (`deleted_at`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `nav_groups`
--

LOCK TABLES `nav_groups` WRITE;
/*!40000 ALTER TABLE `nav_groups` DISABLE KEYS */;
INSERT INTO `nav_groups` VALUES (1,'2026-11-22 15:03:16.212','2026-11-22 15:03:16.212',NULL,'项目01','project01','System');
/*!40000 ALTER TABLE `nav_groups` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `nav_sites`
--

DROP TABLE IF EXISTS `nav_sites`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `nav_sites` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `name` varchar(20) COLLATE utf8mb4_general_ci NOT NULL COMMENT '''网址名''',
  `nav_group_id` bigint unsigned NOT NULL COMMENT '''网址分组ID''',
  `icon_url` varchar(100) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '''网址Icon''',
  `description` varchar(50) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '''网址描述''',
  `link` varchar(100) COLLATE utf8mb4_general_ci NOT NULL COMMENT '''网址链接''',
  `doc_url` varchar(100) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '''网址文档''',
  `creator` varchar(20) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '''创建人''',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uni_nav_sites_name` (`name`),
  KEY `idx_nav_sites_deleted_at` (`deleted_at`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `nav_sites`
--

LOCK TABLES `nav_sites` WRITE;
/*!40000 ALTER TABLE `nav_sites` DISABLE KEYS */;
INSERT INTO `nav_sites` VALUES (1,'2026-11-22 15:03:16.216','2026-11-22 15:03:16.216',NULL,'jenkins',1,'si:jenkins','project01 Jenkins','http://127.0.0.1:8080/','https://www.jenkins.io/doc/','System');
/*!40000 ALTER TABLE `nav_sites` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `notice_boards`
--

DROP TABLE IF EXISTS `notice_boards`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `notice_boards` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `level` bigint NOT NULL COMMENT '''公告级别''',
  `content` varchar(200) COLLATE utf8mb4_general_ci NOT NULL COMMENT '''公告内容''',
  `creator` varchar(20) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '''创建人''',
  PRIMARY KEY (`id`),
  KEY `idx_notice_boards_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `notice_boards`
--

LOCK TABLES `notice_boards` WRITE;
/*!40000 ALTER TABLE `notice_boards` DISABLE KEYS */;
/*!40000 ALTER TABLE `notice_boards` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `operation_logs`
--

DROP TABLE IF EXISTS `operation_logs`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `operation_logs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `username` varchar(20) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '''用户登录名''',
  `ip` varchar(20) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '''Ip地址''',
  `ip_location` varchar(20) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '''Ip所在地''',
  `method` varchar(20) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '''请求方式''',
  `path` varchar(100) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '''访问路径''',
  `remark` varchar(100) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '''备注''',
  `status` int DEFAULT NULL COMMENT '''响应状态码''',
  `start_time` varchar(2048) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '''发起时间''',
  `time_cost` int DEFAULT NULL COMMENT '''请求耗时(ms)''',
  `user_agent` varchar(2048) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '''浏览器标识''',
  PRIMARY KEY (`id`),
  KEY `idx_operation_logs_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `operation_logs`
--

LOCK TABLES `operation_logs` WRITE;
/*!40000 ALTER TABLE `operation_logs` DISABLE KEYS */;
/*!40000 ALTER TABLE `operation_logs` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `role_menus`
--

DROP TABLE IF EXISTS `role_menus`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `role_menus` (
  `menu_id` bigint unsigned NOT NULL,
  `role_id` bigint unsigned NOT NULL,
  PRIMARY KEY (`menu_id`,`role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `role_menus`
--

LOCK TABLES `role_menus` WRITE;
/*!40000 ALTER TABLE `role_menus` DISABLE KEYS */;
INSERT INTO `role_menus` VALUES (1,1),(2,1),(3,1),(4,1),(4,2),(5,1),(6,1),(7,1),(8,1),(9,1),(10,1),(11,1),(12,1),(13,1),(14,1);
/*!40000 ALTER TABLE `role_menus` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `roles`
--

DROP TABLE IF EXISTS `roles`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `roles` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `name` varchar(20) COLLATE utf8mb4_general_ci NOT NULL,
  `keyword` varchar(20) COLLATE utf8mb4_general_ci NOT NULL,
  `remark` varchar(100) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '''备注''',
  `status` tinyint(1) DEFAULT '1' COMMENT '''1正常, 2禁用''',
  `sort` int DEFAULT '999' COMMENT '''角色排序(排序越大权限越低, 不能查看比自己序号小的角色, 不能编辑同序号用户权限, 排序为1表示超级管理员)''',
  `creator` varchar(20) COLLATE utf8mb4_general_ci DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uni_roles_name` (`name`),
  UNIQUE KEY `uni_roles_keyword` (`keyword`),
  KEY `idx_roles_deleted_at` (`deleted_at`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `roles`
--

LOCK TABLES `roles` WRITE;
/*!40000 ALTER TABLE `roles` DISABLE KEYS */;
INSERT INTO `roles` VALUES (1,'2026-11-22 15:03:16.166','2026-11-22 15:03:16.166',NULL,'SuperAdmin','superadmin','',1,1,'System'),(2,'2026-11-22 15:03:16.166','2026-11-22 15:03:16.166',NULL,'Users','user','',1,3,'System'),(3,'2026-11-22 15:03:16.166','2026-11-22 15:03:16.166',NULL,'Guests','guest','',1,5,'System'),(4,'2026-11-22 15:03:16.166','2026-11-22 15:03:16.166',NULL,'BindDN','binddn','',1,5,'System');
/*!40000 ALTER TABLE `roles` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `totps`
--

DROP TABLE IF EXISTS `totps`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `totps` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `user_id` bigint unsigned NOT NULL COMMENT '''用户id''',
  `secret` varchar(32) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '''totp secret''',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uni_totps_secret` (`secret`),
  KEY `idx_totps_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `totps`
--

LOCK TABLES `totps` WRITE;
/*!40000 ALTER TABLE `totps` DISABLE KEYS */;
/*!40000 ALTER TABLE `totps` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `user_roles`
--

DROP TABLE IF EXISTS `user_roles`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `user_roles` (
  `role_id` bigint unsigned NOT NULL,
  `user_id` bigint unsigned NOT NULL,
  PRIMARY KEY (`role_id`,`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `user_roles`
--

LOCK TABLES `user_roles` WRITE;
/*!40000 ALTER TABLE `user_roles` DISABLE KEYS */;
INSERT INTO `user_roles` VALUES (1,1);
/*!40000 ALTER TABLE `user_roles` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `users`
--

DROP TABLE IF EXISTS `users`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `users` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `username` varchar(50) COLLATE utf8mb4_general_ci NOT NULL COMMENT '''用户名''',
  `password` varchar(255) COLLATE utf8mb4_general_ci NOT NULL COMMENT '''用户密码''',
  `nickname` varchar(50) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '''中文名''',
  `given_name` varchar(50) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '''花名''',
  `mail` varchar(100) COLLATE utf8mb4_general_ci NOT NULL COMMENT '''邮箱''',
  `job_number` varchar(20) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '''工号''',
  `mobile` varchar(15) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '''手机号''',
  `avatar` varchar(255) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '''头像''',
  `postal_address` varchar(255) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '''地址''',
  `position` varchar(128) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '''职位''',
  `introduction` varchar(255) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '''个人简介''',
  `status` tinyint(1) DEFAULT '1' COMMENT '''状态:1在职, 2离职''',
  `creator` varchar(20) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '''创建者''',
  `source` varchar(50) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '''用户来源：dingTalk、wecom、feishu、ldap、platform''',
  `source_user_id` varchar(100) COLLATE utf8mb4_general_ci NOT NULL COMMENT '''第三方用户id''',
  `source_union_id` varchar(100) COLLATE utf8mb4_general_ci NOT NULL COMMENT '''第三方唯一unionId''',
  `user_dn` varchar(255) COLLATE utf8mb4_general_ci NOT NULL COMMENT '''用户dn''',
  `sync_state` tinyint(1) DEFAULT '1' COMMENT '''同步状态:1已同步, 2未同步''',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uni_users_username` (`username`),
  UNIQUE KEY `uni_users_mail` (`mail`),
  KEY `idx_users_deleted_at` (`deleted_at`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `users`
--

LOCK TABLES `users` WRITE;
/*!40000 ALTER TABLE `users` DISABLE KEYS */;
INSERT INTO `users` VALUES (1,'2026-11-22 15:03:16.187','2026-11-22 15:03:16.187',NULL,'admin','KRedEx1ed2QAtUpgOQP8f1adxh3VYe4aHnIAarJfyae1d4hWX1uTi9rXvi8RA6UF17JgfYrH1LSFUUCZa3jIeggDCdpzUTkNxqFXvDvVSDbpcSxe0i7GnTH3H3mdVxCS1NgGggHvWAp+EfJBNhk8m0A8o9ZE1bw9miHmLu1nNm8=','Super Admin','Super Admin','admin@example.com','0000','18888888888','https://q1.qlogo.cn/g?b=qq&nk=10002&s=100','default','default','default',1,'System','','','','cn=admin,dc=example,dc=com',1);
/*!40000 ALTER TABLE `users` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2026-11-22 11:03:39
