export default `# delete lines you don't need!
[Unit]
Description=[OpenConMan]
Documentation=nowhere.yet
StartLimitIntervalSec=60
StartLimitBurst=3

[Service]
WorkingDirectory=/etc/conman.d/
EnvironmentFile=/etc/conman.d/conman.env
User= < USER >
Group= < GROUP >
ExecStart= < EXEC >
ExecReload=/bin/kill --signal HUP $MAINPID
KillMode=process
KillSignal=SIGINT
Restart=on-failure
RestartSec=5
TimeoutStopSec=30
LimitMEMLOCK=infinity

[Install]
WantedBy=multi-user.target
`;
