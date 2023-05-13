-- CREATE TABLE `users`
-- (
--     id   bigint auto_increment,
--     name varchar(255) NOT NULL,
--     PRIMARY KEY (`id`)
-- );

-- INSERT INTO `users` (`name`)
-- VALUES ('Solomon'),
--        ('Menelik');

DROP TABLE IF EXISTS io;
CREATE TABLE io (
  id         INT AUTO_INCREMENT NOT NULL,
  mobile      VARCHAR(11) NOT NULL,
  wechat     VARCHAR(50) NOT NULL,
  json_data  json NOT NULL,
  PRIMARY KEY (`id`)
);

DROP TABLE IF EXISTS la;
CREATE TABLE la (
  id         INT AUTO_INCREMENT NOT NULL,
  mobile      VARCHAR(11) NOT NULL,
  wechat     VARCHAR(50) NOT NULL,
  json_data  json NOT NULL,
  PRIMARY KEY (`id`)
);