create database toughradius DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci;
GRANT ALL ON toughradius.* TO raduser@'127.0.0.1' IDENTIFIED BY 'radpwd' WITH GRANT OPTION;FLUSH PRIVILEGES;
# GRANT ALL ON toughradius.* TO raduser@'%' IDENTIFIED BY 'radpwd' WITH GRANT OPTION;FLUSH PRIVILEGES;