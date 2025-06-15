@echo off

pushd %~dp0

setlocal

set output=gh2
set dist=dist
if not exist %dist% mkdir %dist%

go build -trimpath -ldflags="-s -w" -o %dist%\%output%.exe

popd
