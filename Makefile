venv:
	(\
	test -d venv || virtualenv venv;\
	venv/bin/pip install -U pip;\
	venv/bin/pip install -U wheel;\
	venv/bin/pip install -U -r requirements.txt;\
	)


initdb:
	venv/bin/python radiusctl initdb -f -c etc/toughradius.json

clean:
	rm -fr venv

run:
	venv/bin/python radiusctl standalone -c etc/toughradius.json

suprun:
	venv/bin/python radiusctl daemon -s startup -n -c etc/toughradius_test.conf


.PHONY:  venv uplibs test initdb clean run suprun

