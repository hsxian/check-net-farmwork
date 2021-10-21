@echo off
rmdir /s/q publish
go build -o publish/check-runtime-environment.exe  main.go
copy .\config-check-runtime-environment.json .\publish\config-check-runtime-environment.json
pause