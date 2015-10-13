@echo off
setlocal
set GOPATH=%~dp0
go get -u github.com/disintegration/gift
pushd %~dp0src\lec3-ip
go install
popd
endlocal