-- MySQL dump 10.13  Distrib 8.0.45, for Linux (x86_64)
--
-- Host: localhost    Database: test
-- ------------------------------------------------------
-- Server version	8.0.45

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
-- Table structure for table `Abc`
--

DROP TABLE IF EXISTS `Abc`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `Abc` (
  `ID` bigint NOT NULL AUTO_INCREMENT,
  `Name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `Active`
--

DROP TABLE IF EXISTS `Active`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `Active` (
  `ID` bigint NOT NULL AUTO_INCREMENT,
  `UserID` bigint NOT NULL COMMENT '用户ID',
  `LastActiveTime` datetime(3) DEFAULT NULL,
  `Days` date DEFAULT NULL,
  `DaysActiveCount` bigint DEFAULT NULL,
  PRIMARY KEY (`ID`),
  UNIQUE KEY `uk_active_days_userid` (`Days`,`UserID`),
  KEY `idx_Active_user_id` (`UserID`),
  CONSTRAINT `uk_active` FOREIGN KEY (`UserID`) REFERENCES `UserV1` (`ID`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `CollectedVideo`
--

DROP TABLE IF EXISTS `CollectedVideo`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `CollectedVideo` (
  `ID` bigint NOT NULL AUTO_INCREMENT,
  `UserID` bigint NOT NULL COMMENT '用户ID',
  `VideoID` bigint NOT NULL COMMENT '视频ID',
  `CreatedAt` datetime DEFAULT NULL,
  PRIMARY KEY (`ID`),
  KEY `idx_CollectedVideo_user_id` (`UserID`),
  KEY `idx_CollectedVideo_video_id` (`VideoID`),
  CONSTRAINT `uk_user` FOREIGN KEY (`UserID`) REFERENCES `UserV1` (`ID`) ON DELETE CASCADE,
  CONSTRAINT `uk_video` FOREIGN KEY (`VideoID`) REFERENCES `Video` (`ID`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=9 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `Comment`
--

DROP TABLE IF EXISTS `Comment`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `Comment` (
  `ID` bigint NOT NULL AUTO_INCREMENT,
  `ParentID` bigint DEFAULT NULL COMMENT '父视频ID（关联视频/系列）',
  `ChildrenCount` bigint DEFAULT NULL COMMENT '计数',
  `VideoID` bigint NOT NULL COMMENT '视频ID',
  `UserID` bigint DEFAULT NULL,
  `UserID2` bigint DEFAULT NULL COMMENT '评论对谁说',
  `Comment` varchar(2048) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `CreatedAt` datetime(3) DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`ID`),
  KEY `idx_Comment_parent_id` (`ParentID`),
  KEY `idx_Comment_video_id` (`VideoID`),
  KEY `idx_Comment_user_id` (`UserID`),
  CONSTRAINT `uk_comment_parent` FOREIGN KEY (`ParentID`) REFERENCES `Comment` (`ID`) ON DELETE CASCADE,
  CONSTRAINT `uk_user_comment` FOREIGN KEY (`UserID`) REFERENCES `UserV1` (`ID`) ON DELETE SET NULL,
  CONSTRAINT `uk_video_comment` FOREIGN KEY (`VideoID`) REFERENCES `Video` (`ID`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=18 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `Def`
--

DROP TABLE IF EXISTS `Def`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `Def` (
  `ID` bigint NOT NULL AUTO_INCREMENT,
  `AbcID` bigint NOT NULL COMMENT 'AbcID',
  `DefName` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`ID`),
  KEY `idx_Def_abc_id` (`AbcID`),
  CONSTRAINT `fk_abc_def` FOREIGN KEY (`AbcID`) REFERENCES `Abc` (`ID`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `LoginLocation`
--

DROP TABLE IF EXISTS `LoginLocation`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `LoginLocation` (
  `ID` bigint NOT NULL AUTO_INCREMENT,
  `UserID` bigint NOT NULL COMMENT '用户ID',
  `Location` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `CodeCrypto` char(60) COLLATE utf8mb4_unicode_ci NOT NULL,
  `LoginTime` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`ID`),
  KEY `idx_LoginLocation_user_id` (`UserID`),
  CONSTRAINT `uk_loginLocation` FOREIGN KEY (`UserID`) REFERENCES `UserV1` (`ID`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=19 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `Series`
--

DROP TABLE IF EXISTS `Series`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `Series` (
  `ID` bigint NOT NULL AUTO_INCREMENT,
  `ParentID` bigint DEFAULT NULL COMMENT '如果还有上层结构',
  `Layer` bigint NOT NULL DEFAULT '0',
  `BaseOn` enum('Series','Video') COLLATE utf8mb4_unicode_ci DEFAULT 'Video',
  `Name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '系列名称',
  `Description` text COLLATE utf8mb4_unicode_ci,
  `VideosCount` bigint DEFAULT NULL COMMENT '计数',
  `ViewsCount` int DEFAULT NULL,
  `UserID` bigint DEFAULT NULL COMMENT '用户ID',
  `DeletedAt` datetime DEFAULT NULL,
  PRIMARY KEY (`ID`),
  KEY `idx_Series_parent_id` (`ParentID`),
  KEY `idx_Series_name` (`Name`),
  KEY `idx_Series_user_id` (`UserID`),
  KEY `idx_user_deleted_at` (`DeletedAt`),
  CONSTRAINT `cas_parent` FOREIGN KEY (`ParentID`) REFERENCES `Series` (`ID`) ON DELETE CASCADE,
  CONSTRAINT `uk_series_user` FOREIGN KEY (`UserID`) REFERENCES `UserV1` (`ID`) ON DELETE SET NULL
) ENGINE=InnoDB AUTO_INCREMENT=49 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `UserFollow`
--

DROP TABLE IF EXISTS `UserFollow`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `UserFollow` (
  `ID` bigint NOT NULL AUTO_INCREMENT,
  `FollowerID` bigint NOT NULL,
  `FolloweeID` bigint NOT NULL,
  `CreatedAt` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`ID`),
  UNIQUE KEY `uk_follow` (`FollowerID`,`FolloweeID`),
  KEY `uk_followee` (`FolloweeID`),
  CONSTRAINT `uk_followee` FOREIGN KEY (`FolloweeID`) REFERENCES `UserV1` (`ID`) ON DELETE CASCADE,
  CONSTRAINT `uk_follower` FOREIGN KEY (`FollowerID`) REFERENCES `UserV1` (`ID`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=11 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `UserV1`
--

DROP TABLE IF EXISTS `UserV1`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `UserV1` (
  `ID` bigint NOT NULL AUTO_INCREMENT,
  `Name` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `PasswordCrypto` char(60) COLLATE utf8mb4_unicode_ci NOT NULL,
  `PhoneCipherDB` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '手机号(含国家前缀)',
  `PhoneHash` binary(32) DEFAULT NULL,
  `EmailCipherDB` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `EmailHash` binary(32) DEFAULT NULL,
  `Status` enum('happy','sad','anxious','angry','calm') COLLATE utf8mb4_unicode_ci DEFAULT 'happy',
  `Bytes` json DEFAULT NULL,
  `CreatedAt` datetime(3) DEFAULT NULL,
  `FollowingCount` bigint DEFAULT '0',
  `FollowerCount` bigint DEFAULT '0',
  PRIMARY KEY (`ID`),
  UNIQUE KEY `idx_UserV1_name` (`Name`),
  UNIQUE KEY `idx_user_email` (`EmailHash`),
  UNIQUE KEY `idx_user_phone` (`PhoneHash`),
  KEY `idx_UserV1_phone_hash` (`PhoneHash`),
  KEY `idx_UserV1_email_hash` (`EmailHash`),
  KEY `idx_user_name` (`Name`)
) ENGINE=InnoDB AUTO_INCREMENT=28 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `Video`
--

DROP TABLE IF EXISTS `Video`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `Video` (
  `SeriesID` bigint DEFAULT NULL COMMENT '父id,系列',
  `ID` bigint NOT NULL AUTO_INCREMENT COMMENT '视频唯一ID',
  `UserID` bigint DEFAULT NULL COMMENT '用户ID',
  `Name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '视频名称',
  `Duration` double DEFAULT NULL COMMENT '时长',
  `Description` text COLLATE utf8mb4_unicode_ci,
  `Path` varchar(500) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '所谓的视频所在地',
  `ViewsCount` bigint NOT NULL DEFAULT '0',
  `Tags` json DEFAULT NULL COMMENT '标签JSON数组',
  `CreatedAt` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `CommentsCount` bigint NOT NULL DEFAULT '0',
  `CollectedVideosCount` bigint DEFAULT '0',
  PRIMARY KEY (`ID`),
  KEY `idx_Video_series_id` (`SeriesID`),
  KEY `idx_Video_user_id` (`UserID`),
  KEY `idx_Video_name` (`Name`),
  CONSTRAINT `cas_seriesID` FOREIGN KEY (`SeriesID`) REFERENCES `Series` (`ID`) ON DELETE CASCADE,
  CONSTRAINT `userv1_video` FOREIGN KEY (`UserID`) REFERENCES `UserV1` (`ID`) ON DELETE SET NULL
) ENGINE=InnoDB AUTO_INCREMENT=55 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `View`
--

DROP TABLE IF EXISTS `View`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `View` (
  `ID` bigint NOT NULL AUTO_INCREMENT,
  `VideoID` bigint NOT NULL COMMENT '视频ID',
  `UserID` bigint NOT NULL COMMENT '用户ID',
  `LastViewTime` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`ID`),
  UNIQUE KEY `uk_video_user` (`VideoID`,`UserID`),
  KEY `idx_View_user_id` (`UserID`),
  CONSTRAINT `uk_video_view` FOREIGN KEY (`VideoID`) REFERENCES `Video` (`ID`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=17 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping events for database 'test'
--

--
-- Dumping routines for database 'test'
--
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2026-10-30 12:36:17
