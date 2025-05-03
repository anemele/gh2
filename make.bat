@echo off

setlocal

set output=ghdl
set dist=dist
if not exist %dist% mkdir %dist%

go build -ldflags="-s -w" -o %dist%\ghdl.exe
