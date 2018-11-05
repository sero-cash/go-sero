#  go-sero RPM package generate and installation/uninstallation

## centos 7

from the cloned resource github/sero-cash/go-sero
goto build directory
run 
```
./rpmbuild.sh
```

it will generate rpm package in github/sero-cash/go-sero/build/package/RPMS

## install with rpm
```
rpm -ivh ${rpmfile} --nodeps
```


## check install and env
exec following command in console
```
gero
```
it it complains with missing libboost_system ... etc.
please ref [installation guide](https://github.com/sero-cash/go-sero/wiki/Building-Sero)

for centos
```
sudo yum --setopt=group_package_types=mandatory,default,optional group install "Development Tools"
sudo yum install boost boost-devel boost-system boost-filesystem boost-thread
```
for ubuntu
```
sudo apt-get install -y build-essential golang
sudo apt-get install libboost-all-dev
```

## uninstall
if you met following error:

  install of gero-1.0-1.x86_64 conflicts with file from package gero-1.0-1.x86_64
or  you want to upgrade gero
you need to uninstall previous gero package

```
sudo dpkg --purge ${packagename}
```
or 
```
sudo rpm -e ${packagename}
```