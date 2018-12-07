@echo off 
set CURRENT=%cd%
echo %CURRENT%
set LIB_PATH=%CURRENT%\czero\lib
echo %LIB_PATH%
set path=%LIB_PATH%

if  not exist %CURRENT%\log (
     md %CURRENT%\log
)

C:\Windows\system32\taskkill.exe /f /im gero.exe 
start %CURRENT%\bin\gero.exe >%CURRENT%\log\log.txt

