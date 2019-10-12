CREATE TABLE `link`
(
    `short` varchar(64) COLLATE utf8mb4_unicode_ci   NOT NULL,
    `long`  varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL,
    `user`  varchar(20) COLLATE utf8mb4_unicode_ci   NOT NULL,
    PRIMARY KEY (`short`),
    UNIQUE KEY `short` (`short`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci