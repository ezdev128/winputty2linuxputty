# winputty2linuxputty
---
Utility to transform PuTTY sessions from Windows Registry file into Linux PuTTY sessions.

## Build:
go get

go build -o putty_sessions_converter .

## Usage:
### Prepare sessions dump:
reg export HKEY_CURRENT_USER\Software\SimonTatham\PuTTY\Sessions C:\putty_sessions.txt

### Convert registry dump to sessions dir:

putty_sessions_converter C:\putty_sessions.txt 
