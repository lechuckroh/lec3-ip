@echo off
setlocal
set GOPATH=%~dp0
go get -u github.com/disintegration/gift
go get -u github.com/olebedev/config
go get -u github.com/mitchellh/mapstructure
pushd %~dp0src\lec3-ip
go install
popd
endlocal