#!/bin/sh
currdir=$PWD
version="v6.1.1.4"
releasedir=${currdir}/release/toughradius-${version}
releasefile=toughradius-${version}.zip


build_version()
{
    echo "release version ${version}"
    test -d ${releasedir} || mkdir ${releasedir}
    rm -fr ${releasedir}/*
    test -f ${releasefile} && rm -f ${releasefile}

    cp -r ${currdir}/src/main/resources/portal ${releasedir}
    cp ${currdir}/scripts/application-prod.properties ${releasedir}/application-prod.properties
    cp ${currdir}/scripts/freeradius-rest.conf ${releasedir}/freeradius-rest.conf
    cp ${currdir}/scripts/createdb.sql ${releasedir}/createdb.sql
    cp ${currdir}/scripts/database.sql ${releasedir}/database.sql
    cp ${currdir}/scripts/init.sql ${releasedir}/init.sql
    cp ${currdir}/scripts/installer.sh ${releasedir}/installer.sh
    cp ${currdir}/scripts/toughradius.service ${releasedir}/toughradius.service
    cp ${currdir}/scripts/linux-installer.md ${releasedir}/linux-installer.md
    cp ${currdir}/README.md ${releasedir}/README.md
    dos2unix ${releasedir}/*.properties
    dos2unix ${releasedir}/*.sql
    dos2unix ${releasedir}/*.sh
    dos2unix ${releasedir}/*.service
    cp ${currdir}/scripts/startup.bat ${releasedir}/startup.bat
    cp ${currdir}/target/toughradius-latest.jar ${releasedir}/toughradius-latest.jar
    cd ${currdir}/release && zip -r ${releasefile} toughradius-${version}
    echo "release file ${releasefile}"
}


case "$1" in

  build)
    build_version
  ;;

esac