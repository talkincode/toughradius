BUILD_ORG   := talkincode
BUILD_VERSION   := latest
BUILD_TIME      := $(shell date "+%F %T")
BUILD_NAME      := toughradius
RELEASE_VERSION := v8.0.7
SOURCE          := main.go
RELEASE_DIR     := ./release
COMMIT_SHA1     := $(shell git show -s --format=%H )
COMMIT_DATE     := $(shell git show -s --format=%cD )
COMMIT_USER     := $(shell git show -s --format=%ce )
COMMIT_SUBJECT     := $(shell git show -s --format=%s )

buildpre:
	echo "BuildVersion=${BUILD_VERSION} ${RELEASE_VERSION} ${BUILD_TIME}" > assets/buildinfo.txt
	echo "ReleaseVersion=${RELEASE_VERSION}" >> assets/buildinfo.txt
	echo "BuildTime=${BUILD_TIME}" >> assets/buildinfo.txt
	echo "BuildName=${BUILD_NAME}" >> assets/buildinfo.txt
	echo "CommitID=${COMMIT_SHA1}" >> assets/buildinfo.txt
	echo "CommitDate=${COMMIT_DATE}" >> assets/buildinfo.txt
	echo "CommitUser=${COMMIT_USER}" >> assets/buildinfo.txt
	echo "CommitSubject=${COMMIT_SUBJECT}" >> assets/buildinfo.txt

build:
	test -d ./release || mkdir -p ./release
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags  '-s -w -extldflags "-static"'  -o ./release/toughradius main.go
	upx ./release/toughradius

buildarm64:
	test -d ./release || mkdir -p ./release
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -a -ldflags  '-s -w -extldflags "-static"'  -o ./release/toughradius main.go
	upx ./release/toughradius

build-tradtest:
	CGO_ENABLED=0 go build -a -ldflags '-s -w -extldflags "-static"' -o release/bmtest commands/benchmark/bmtest.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-s -w -extldflags "-static"' -o release/lbmtest commands/benchmark/bmtest.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-s -w -extldflags "-static"' -o release/bmtest.exe commands/benchmark/bmtest.go


syncdev:
	make buildpre
	@read -p "提示:同步操作尽量在完成一个完整功能特性后进行，请输入提交描述 (develop):  " cimsg; \
	git commit -am "$(shell date "+%F %T") : $${cimsg}"
	# 切换主分支并更新
	git checkout main
	git pull origin main
	# 切换开发分支变基合并提交
	git checkout develop
	git rebase -i main
	# 切换回主分支并合并开发者分支，推送主分支到远程，方便其他开发者合并
	git checkout main
	git merge --no-ff develop
	git push origin main
	# 切换回自己的开发分支继续工作
	git checkout develop

tr069crt:
	# 1 Generate CA private key
	test -f assets/ca.key || openssl genrsa -out assets/ca.key 4096
	# 2 Generate CA certificate
	test -f assets/ca.crt || openssl req -x509 -new -nodes -key assets/ca.key -days 3650 -out assets/ca.crt -subj \
	"/C=CN/ST=Shanghai/O=toughradius/CN=ToughradiusCA/emailAddress=master@toughstruct.net"
	# 3 Generate server private key
	openssl genrsa -out assets/server.key 2048
	# 4 Generate a certificate request file
	openssl req -new -key assets/server.key -out assets/server.csr -subj \
	"/C=CN/ST=Shanghai/O=toughradius/CN=*.toughstruct.net/emailAddress=master@toughstruct.net"
	# 5 Generate a server certificate based on the CA's private key and the above certificate request file
	openssl x509 -req -in assets/server.csr -CA assets/ca.crt -CAkey assets/ca.key -CAcreateserial -out assets/server.crt -days 7300
	mv assets/server.key assets/cwmp.tls.key
	mv assets/server.crt assets/cwmp.tls.crt

radseccrt:
	# 1 Generate CA private key
	test -f assets/ca.key || openssl genrsa -out assets/ca.key 4096
	# 2 Generate CA certificate
	test -f assets/ca.crt || openssl req -x509 -new -nodes -key assets/ca.key -days 3650 -out assets/ca.crt -subj \
	"/C=CN/ST=Shanghai/O=toughradius/CN=ToughradiusCA/emailAddress=master@toughstruct.net"
	# 3 Generate server private key
	openssl genrsa -out assets/server.key 2048
	# 4 Generate a certificate request file
	openssl req -new -key assets/server.key -out assets/server.csr -subj \
	"/C=CN/ST=Shanghai/O=toughradius/CN=*.toughstruct.net/emailAddress=master@toughstruct.net"
	# 5 Generate a server certificate based on the CA's private key and the above certificate request file
	openssl x509 -req -in assets/server.csr -CA assets/ca.crt -CAkey assets/ca.key -CAcreateserial -out assets/server.crt -days 7300
	mv assets/server.key assets/radsec.tls.key
	mv assets/server.crt assets/radsec.tls.crt

clicrt:
	# 1 生成client私钥
	openssl genrsa -out assets/client.key 2048
	# 2 生成client请求文件
	openssl req -new -key assets/client.key -subj "/CN=*.toughstruct.net" -out assets/client.csr
	# 3 生成client证书
	openssl x509 -req -in assets/client.csr -CA assets/ca.crt -CAkey assets/ca.key -CAcreateserial -out assets/client.crt -days 7300
	mv assets/client.key assets/client.tls.key
	mv assets/client.crt assets/client.tls.crt

updev:
	make buildpre
	make build
	scp ${RELEASE_DIR}/${BUILD_NAME} trdev-server:/tmp/toughradius
	ssh trdev-server "systemctl stop toughradius && /tmp/toughradius -install && systemctl start toughradius"

swag:
	swag fmt && swag init


syncdev:
	@echo "🚀 开始执行同步流程（develop → main）..."
	@./scripts/syncdev.sh

tag:
	@echo "🏷️  开始标签创建流程..."
	@./scripts/tag.sh

.PHONY: clean build tr069crt radseccrt

