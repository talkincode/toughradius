
#生成服务器端证书

keytool -genkey -keyalg RSA -dname “cn=toughstruct,ou=toughstruct,o=toughstruct,l=cs,st=hn,c=cn” -alias server -storetype PKCS12 -keypass radsec -keystore server.jks -storepass radsec -validity 3650



# 生成客户端证书

keytool -genkey -keyalg RSA -dname “cn=toughstruct,ou=toughstruct,o=toughstruct,l=cs,st=hn,c=cn” -alias client -storetype PKCS12 -keypass radsec -keystore client.p12 -storepass radsec -validity 3650

keytool -export -alias client -file client.cer -keystore client.p12 -storepass radsec -storetype PKCS12 -rfc


#添加客户端证书到服务器中（将已签名数字证书导入密钥库）

keytool -import -v -alias client -file client.cer -keystore server.jks -storepass radsec








