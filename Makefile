venv:
	(\
	test -d venv || virtualenv venv;\
	venv/bin/pip install -U pip;\
	venv/bin/pip install -U wheel;\
	venv/bin/pip install -U -r requirements.txt;\
	)

venv2:
	(\
	test -d venv || virtualenv-pypy venv;\
	venv/bin/pip install gevent;\
	)


install:
	venv/bin/python setup.py install

bdist:
	venv/bin/python setup.py bdist

wheel:
	venv/bin/python setup.py bdist_wheel

rpm:
	venv/bin/python setup.py bdist_rpm

upload:
	venv/bin/python setup.py bdist_rpm bdist_wheel upload

doc:
	cd documents && sphinx-intl update -p build/locale -l zh_CN
	cd documents && sphinx-intl build && make -e SPHINXOPTS="-D language='zh_CN'" html
	rm -fr /docs/*
	rsync -av documents/build/html/ docs/
	rm -fr documents/build/html
	echo "docs.toughradius.net" > docs/CNAME
	echo "" > docs/.nojekyll
	venv/bin/python documents/rename.py

test:
	echo 'test'

clean:
	rm -fr toughradius.egg-info
	rm -fr dist
	rm -fr build
	rm -fr venv

run:
	venv/bin/python radiusd.py -x


.PHONY:  venv test  clean 

