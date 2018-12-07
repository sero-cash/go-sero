@echo off 
set CURRENT=%cd%
echo %CURRENT%
set LIB_PATH=%CURRENT%\czero\lib
echo %LIB_PATH%
set path=%LIB_PATH%
start /b %CURRENT%\bin\gero.exe attach \\.\pipe\gero.ipc

