CREATE TABLE `link`
(
    `short`     varchar(64) COLLATE utf8mb4_unicode_ci   NOT NULL,
    `long`      varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL,
    `user`      varchar(36) COLLATE utf8mb4_unicode_ci   NOT NULL,
    `user_name` varchar(50) collate utf8mb4_unicode_ci   not null,
    PRIMARY KEY (`short`),
    UNIQUE KEY `short` (`short`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci;

CREATE TABLE `unsplash_image`
(
    `id`                   int(11)                                 NOT NULL,
    `url`                  varchar(200) COLLATE utf8mb4_unicode_ci NOT NULL,
    `photographer_name`    varchar(50) COLLATE utf8mb4_unicode_ci  NOT NULL,
    `photographer_profile` varchar(50) COLLATE utf8mb4_unicode_ci  NOT NULL,
    `updated`              datetime                                NOT NULL,
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci;

INSERT INTO `unsplash_image` (`id`, `url`, `photographer_name`, `photographer_profile`, `updated`)
VALUES (1, '', '', '', NOW());