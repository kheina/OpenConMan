package daemon

import (
	"fmt"
	"os"
	"strings"

	"github.com/kheina/openconman/src/util"
)

const unitfiletemplate = `[Unit]
Description=OpenConMan
Documentation=nowhere.yet
StartLimitIntervalSec=60
StartLimitBurst=3

[Service]
WorkingDirectory=/etc/conman.d/
EnvironmentFile=/etc/conman.d/conman.env
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
const workingDirectory = "/etc/conman.d/"
const environmentFile = "/etc/conman.d/conman.env"

func newUnitFile(exe, filename string, args []string) (string, error) {
	contents := fmt.Sprintf(unitfiletemplate, exe, strings.Join(args, " "))
	path := fmt.Sprintf("/etc/systemd/system/%s", filename)
	// r = 100b = 4
	// w = 010b = 2
	// x = 001b = 1
	// 644 == -rw-r--r--
	if err := os.WriteFile(path, []byte(contents), 0644); err != nil {
		return "", err
	}
	if !util.PathExists(workingDirectory) {
		if err := os.Mkdir(workingDirectory, 0644); err != nil {
			return "", err
		}
	}
	if !util.PathExists(environmentFile) {
		if err := os.WriteFile(environmentFile, []byte{}, 0644); err != nil {
			return "", err
		}
	}
	return path, nil
}

func deleteUnitFile(filename string) error {
	return os.Remove(fmt.Sprintf("/etc/systemd/system/%s", filename))
}
