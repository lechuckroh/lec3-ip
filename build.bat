@echo off
setlocal
set GOPATH=%~dp0
pushd %~dp0src\lec3-ip
go install
popd
endlocal