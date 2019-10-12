create table link
(
    short  varchar(64)   NOT NULL,
    `long` varchar(1024) NOT NULL,

    primary key (short),
    unique index (short)
)