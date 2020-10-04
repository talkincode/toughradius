FROM openjdk:8

USER root

COPY toughradius.db.sql /toughradius.sql
COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

RUN curl -k https://estatic.toughstruct.net/toughradius-latest.jar -o /toughradius.jar

ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["java", "-jar", "/toughradius.jar"]