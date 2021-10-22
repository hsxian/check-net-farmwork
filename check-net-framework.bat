@echo off
reg query "HKLM\Software\Microsoft\NET Framework Setup\NDP" /s /v version | findstr /i version | sort /+26 /r 
pause