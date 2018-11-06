#!/bin/bash
START_PATH=`pwd`
SOURCE_PATH=`pwd`/../../
NAME_PATTERN="Name:"
VERSION_PATTERN="Version:"
while read line; do
    if [[ $line =~ ${NAME_PATTERN} ]]; then
        fields=(`echo $line`);
        NAME=${fields[1]}
        echo "NAME:${NAME}"
    elif [[ $line =~ ${VERSION_PATTERN} ]]; then
        fields=(`echo $line`);
        VERSION=${fields[1]}
        echo "VERSION:${VERSION}"
    fi
done <${START_PATH}/gero.spec

if [ ! -d ./package ]; then
    mkdir ./package
else
    rm -rf ./package
    mkdir ./package
fi
if [ ! -d ./tmp ]; then
    mkdir ./tmp 
else
    rm -rf ./tmp
    mkdir ./tmp
fi


if [ ! -f ${START_PATH}/bin/gero ]; then
    cd ${START_PATH}/../
    make clean 
    make all
fi


if [ ! -d ${SOURCE_PATH}/go-czero-import ]; then
    echo "there is no project available for package"
    exit 1
fi
cd ${START_PATH}/package
mkdir -p czero
mkdir -p bin
cp -rf ${START_PATH}/bin/* ./bin/
mv  ./bin/gero ./bin/_gero
cp -rf ${SOURCE_PATH}/go-czero-import/czero/* ./czero/
#cat > ${START_PATH}/package/gero << EOL
##!/bin/bash
#cd "$(dirname $BASH_SOURCE)"
#CUR_PATH=`pwd`
#export DYLD_LIBRARY_PATH=${CUR_PATH}/czero/lib
#${CUR_PATH}/bin/_gero $1
#EOL
cat "" > ${START_PATH}/package/gero 
echo '#!/bin/bash' >>${START_PATH}/package/gero  
echo 'cd "$(dirname $BASH_SOURCE)"' >>${START_PATH}/package/gero  
echo 'CUR_PATH=`pwd`' >>${START_PATH}/package/gero  
echo 'export DYLD_LIBRARY_PATH=${CUR_PATH}/czero/lib' >>${START_PATH}/package/gero  
echo '${CUR_PATH}/bin/_gero $1' >>${START_PATH}/package/gero  
chmod -R 755 *
cd ${START_PATH}
cd ${START_PATH}/tmp
hdiutil create ${START_PATH}/tmp.dmg -ov -volname "Gero_Mac_Instl" -fs HFS+ -srcfolder "${START_PATH}/package"
hdiutil convert ${START_PATH}/tmp.dmg -format UDZO -o Gero_Mac_Instl.dmg
