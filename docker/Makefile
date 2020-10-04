build:
	docker build -t talkincode/toughradius .

push:
	docker push talkincode/toughradius

push-github:
	docker tag talkincode/toughradius docker.pkg.github.com/talkincode/toughradius/toughradius
	docker push docker.pkg.github.com/talkincode/toughradius/toughradius

initdb:
	curl -k https://estatic.toughstruct.net/toughradius.db.sql | mysql -uroot -pmyroot -v

cleardb:
	mysql -h 127.0.0.1 -uroot -pmyroot -e "drop database toughradius;"


.PHONY: build


