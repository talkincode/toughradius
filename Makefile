install:
	(\
	virtualenv venv --relocatable;\
	test -d /var/toughradius/data || mkdir -p /var/toughradius/data;\
	rm -f /etc/toughradius.conf && cp etc/toughradius.conf /etc/toughradius.conf;\
	rm -f /etc/toughradius.json && cp etc/toughradius.json /etc/toughradius.json;\
	rm -f /etc/init.d/toughradius && cp etc/toughradius /etc/init.d/toughradius;\
	chmod +x /etc/init.d/toughradius && chkconfig toughradius on;\
	rm -f /usr/lib/systemd/system/toughradius.service && cp etc/toughradius.service /usr/lib/systemd/system/toughradius.service;\
	chmod 754 /usr/lib/systemd/system/toughradius.service && systemctl enable toughradius;\
	systemctl daemon-reload;\
	)

install-deps:
	(\
	yum install -y epel-release;\
	yum install -y wget zip python-devel libffi-devel openssl openssl-devel gcc git;\
	yum install -y czmq czmq-devel python-virtualenv supervisor;\
	yum install -y mysql-devel MySQL-python redis;\
	test -f /usr/local/bin/supervisord || ln -s `which supervisord` /usr/local/bin/supervisord;\
	)

venv:
	(\
	test -d venv || virtualenv venv;\
	venv/bin/pip install -U pip;\
	venv/bin/pip install -U wheel;\
	venv/bin/pip install -U coverage;\
	venv/bin/pip install -U -r requirements.txt;\
	)

upgrade-libs:
	(\
	venv/bin/pip install -U --no-deps https://github.com/talkincode/toughlib/archive/toughlib_cpe.zip;\
	venv/bin/pip install -U --no-deps https://github.com/talkincode/txradius/archive/master.zip;\
	)

upgrade-dev:
	git pull --rebase --stat origin release-dev

upgrade:
	git pull --rebase --stat origin release-stable-cpe

test:
	sh runtests.sh

initdb:
	venv/bin/python radiusctl initdb -f -c /etc/toughradius.json

inittest:
	venv/bin/python radiusctl inittest -c /etc/toughradius.json

clean:
	rm -fr venv

all:install-deps venv upgrade-libs install

.PHONY: all install install-deps upgrade-libs upgrade-dev upgrade test initdb inittest clean