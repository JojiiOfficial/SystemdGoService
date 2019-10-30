# Systemd Go Service
Create, edit, start, stop, enable and disable systemd services easily with Go.

# install
```Go
go get github.com/JojiiOfficial/SystemdGoService
```

# Usage

main.go
```Go
package main

import (
	"github.com/JojiiOfficial/SystemdGoService"
)

func main() {
	service := SystemdGoService.NewDefaultService("testService", "this is a test", "/bin/sh /test.sh")
	service.Service.User = "root"
	service.Service.Group = "root"
	service.Create()
}
``` 
This creates following file (/etc/systemd/system/testService.service):
```
[Unit]
Description=this is a test
After=network.target

[Service]
Type=simple
ExecStart=/bin/sh /test.sh
User=root
Group=root

[Install]
WantedBy=multi-user.target
```
<br>

You can set

For [Unit]
  
- Description
- Documentation
- Before
- After
- Wants
- ConditionPathExisis
- Conflicts
<br>

For [Service]

- Type
- ExecStartPre
- ExecStart
- ExecReload
- ExecStop
- RestartSec
- User
- Group
- Restart
- TimeoutStartSec
- TimeoutStopSec
- SuccessExitStatus
- RestartPreventExitStatus
- PIDFile
- WorkingDirectory
- RootDirectory
- LogsDirectory
- KillMode
- ConditionPathExists
- RemainAfterExit
<br>

For [Install]
- WantedBy
- Alias
