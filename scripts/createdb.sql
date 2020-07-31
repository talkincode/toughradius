create database toughradius DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
GRANT ALL ON toughradius.* TO raduser@'127.0.0.1' IDENTIFIED BY 'radpwd' WITH GRANT OPTION;FLUSH PRIVILEGES;
# GRANT ALL ON toughradius.* TO raduser@'%' IDENTIFIED BY 'radpwd' WITH GRANT OPTION;FLUSH PRIVILEGES;

--  mysql 8
-- create schema toughradius collate utf8mb4_unicode_ci;
-- CREATE USER 'raduser'@'127.0.0.1' identified by 'radpwd';
-- GRANT ALL PRIVILEGES ON toughradius.* TO 'raduser'@'127.0.0.1';