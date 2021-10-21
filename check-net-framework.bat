@echo off
set frameworks_cmd = reg query "HKLM\Software\Microsoft\NET Framework Setup\NDP" /s /v version | findstr /i version | sort /+26 /r 
for /F %%i in ('') do ( set frameworks=%%i)
echo frameworks_cmd=%frameworks_cmd%
pause