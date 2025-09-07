package daemon

import (
	"fmt"
	"os"
	"strings"
)

const unitfiletemplate = `[Unit]
Description=OpenConMan
Documentation=nowhere.yet
StartLimitIntervalSec=60
StartLimitBurst=3

[Service]
WorkingDirectory=/etc/conman.d/
EnvironmentFile=-/etc/conman.d/conman.env
User=root
Group=root
ExecStart=%s serve %s
ExecReload=/bin/kill --signal HUP $MAINPID
KillMode=process
KillSignal=SIGINT
Restart=on-failure
RestartSec=5
TimeoutStopSec=30
LimitMEMLOCK=infinity

[Install]
WantedBy=multi-user.target
`

func newUnitFile(exe, filename string, args []string) error {
	contents := fmt.Sprintf(unitfiletemplate, exe, strings.Join(args, " "))
	// r = 100b = 4
	// w = 010b = 2
	// x = 001b = 1
	// 644 == -rw-r--r--
	return os.WriteFile(fmt.Sprintf("/usr/lib/systemd/system/%s", filename), []byte(contents), 644)
}

func deleteUnitFile(filename string) error {
	return os.Remove(fmt.Sprintf("/usr/lib/systemd/system/%s", filename))
}
