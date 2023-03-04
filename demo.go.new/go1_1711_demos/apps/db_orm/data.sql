
-- CREATE DATABASE IF NOT EXISTS test DEFAULT CHARACTER SET utf8mb4;

CREATE TABLE `user` (
  `id`    int(10) unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
  `name`  varchar(100) NOT NULL COMMENT '名称',
  `age`   int(11) NOT NULL DEFAULT '0' COMMENT '年龄',
  `ctime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `mtime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8mb4;

-- INSERT

INSERT INTO `user` (`name`, `age`, `ctime`, `mtime`)
VALUES ('bar', 29, '2023-03-03 07:59:49', '2023-03-03 07:59:49');
INSERT INTO `user` (`name`, `age`, `ctime`, `mtime`)
VALUES ('foo', 40, '2023-03-03 08:00:03', '2023-03-03 08:00:03');
