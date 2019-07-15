@echo off 
set CURRENT=%cd%
set LIB_PATH=%CURRENT%\czero\lib
set path=%LIB_PATH%
set DATADIR=
set KEYSTORE=
set d=%1
if "%d%" neq "" (
   set DATADIR=--datadir  %d%
)
set k=%2
if "%k%" neq "" (
   set KEYSTORE=--keystore  %k%
)
start /b bin\gero.exe --config geroConfig.toml --exchange --mineMode %DATADIR% %KEYSTORE%

pause

