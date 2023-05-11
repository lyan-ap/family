DROP TABLE IF EXISTS io;
CREATE TABLE io (
  id         INT AUTO_INCREMENT NOT NULL,
  mobile      VARCHAR(11) NOT NULL,
  wechat     VARCHAR(50) NOT NULL,
  json_data  json,
  PRIMARY KEY (`id`)
);

DROP TABLE IF EXISTS la;
CREATE TABLE la (
  id         INT AUTO_INCREMENT NOT NULL,
  mobile      VARCHAR(11) NOT NULL,
  wechat     VARCHAR(50) NOT NULL,
  json_data  json,
  PRIMARY KEY (`id`)
);

-- INSERT INTO `io` (`mobile`,`wechat`)
-- VALUES ('17712345678','wechat123',{}),
--        ('18912345678','wechat456',{});