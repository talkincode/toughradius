venv:
	(\
	test -d venv || virtualenv venv;\
	venv/bin/pip install -U pip;\
	venv/bin/pip install -U wheel;\
	venv/bin/pip install -U -r requirements.txt;\
	)


install:
	python setup.py install

bdist:
	python setup.py bdist

wheel:
	python setup.py bdist_wheel

upload-bdist:
	python setup.py bdist upload

upload-all:
	python setup.py bdist bdist_wheel upload

clean:
	rm -fr toughradius.egg-info
	rm -fr dist
	rm -fr build
	rm -fr venv

run:
	python toughradius/common/commands.py auth -p 10


.PHONY:  venv test  clean 

