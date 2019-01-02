@echo off 
set CURRENT=%cd%
set LIB_PATH=%CURRENT%\czero\lib
set path=%LIB_PATH%
C:\Windows\system32\taskkill.exe /f /im gero.exe
start /b bin\gero.exe

