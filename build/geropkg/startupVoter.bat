@echo off 
set CURRENT=%cd%
set LIB_PATH=%CURRENT%\czero\lib
set path=%LIB_PATH%
set DATADIR=
set KEYSTORE=
set VOTER=
set VOTER_PASSWORD_PATH=


set v=%1
if "%v%" neq "" (
    set VOTER=--unlock %V%
)
set p=%2
if "%v%" neq "" (
    set VOTER_PASSWORD_PATH=--password %p%
)

set d=%3
if "%d%" neq "" (
   set DATADIR=--datadir  %d%
)
set k=%4
if "%k%" neq "" (
   set KEYSTORE=--keystore  %k%
)


start /b bin\gero.exe --config geroConfig.toml --exchange --mineMode %DATADIR% %KEYSTORE% %VOTER% %VOTER_PASSWORD_PATH%

pause

