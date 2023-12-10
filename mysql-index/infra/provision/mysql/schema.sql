CREATE DATABASE IF NOT EXISTS `users`;

GRANT ALL PRIVILEGES ON *.* TO 'root'@'%';

CREATE TABLE if not exists `users`.users
(
    id bigint unsigned auto_increment primary key,
    date_of_birth datetime NOT NULL
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;