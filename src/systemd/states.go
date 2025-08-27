package systemd

type serviceUnitState string

// Available service unit substates
const (
	dead                    serviceUnitState = "dead"
	condition               serviceUnitState = "condition"
	startPre                serviceUnitState = "start-pre"
	start                   serviceUnitState = "start"
	startPost               serviceUnitState = "start-post"
	running                 serviceUnitState = "running"
	exited                  serviceUnitState = "exited"
	reload                  serviceUnitState = "reload"
	reloadSignal            serviceUnitState = "reload-signal"
	reloadNotify            serviceUnitState = "reload-notify"
	mounting                serviceUnitState = "mounting"
	stop                    serviceUnitState = "stop"
	stopWatchdog            serviceUnitState = "stop-watchdog"
	stopSigterm             serviceUnitState = "stop-sigterm"
	stopSigkill             serviceUnitState = "stop-sigkill"
	stopPost                serviceUnitState = "stop-post"
	finalWatchdog           serviceUnitState = "final-watchdog"
	finalSigterm            serviceUnitState = "final-sigterm"
	finalSigkill            serviceUnitState = "final-sigkill"
	failed                  serviceUnitState = "failed"
	deadBeforeAutoRestart   serviceUnitState = "dead-before-auto-restart"
	failedBeforeAutoRestart serviceUnitState = "failed-before-auto-restart"
	deadResourcesPinned     serviceUnitState = "dead-resources-pinned"
	autoRestart             serviceUnitState = "auto-restart"
	autoRestartQueued       serviceUnitState = "auto-restart-queued"
	cleaning                serviceUnitState = "cleaning"
	unknown                 serviceUnitState = "unknown"
)

func allServiceUnitStateStrings() []string {
	return []string{
		string(dead),
		string(condition),
		string(startPre),
		string(start),
		string(startPost),
		string(running),
		string(exited),
		string(reload),
		string(reloadSignal),
		string(reloadNotify),
		string(mounting),
		string(stop),
		string(stopWatchdog),
		string(stopSigterm),
		string(stopSigkill),
		string(stopPost),
		string(finalWatchdog),
		string(finalSigterm),
		string(finalSigkill),
		string(failed),
		string(deadBeforeAutoRestart),
		string(failedBeforeAutoRestart),
		string(deadResourcesPinned),
		string(autoRestart),
		string(autoRestartQueued),
		string(cleaning),
		string(unknown),
	}
}

type unitFileState string

// Available unit file states
const (
	enabled        unitFileState = "enabled"
	enabledRuntime unitFileState = "enabled-runtime"
	linked         unitFileState = "linked"
	linkedRuntime  unitFileState = "linked-runtime"
	alias          unitFileState = "alias"
	masked         unitFileState = "masked"
	maskedRuntime  unitFileState = "masked-runtime"
	static         unitFileState = "static"
	disabled       unitFileState = "disabled"
	indirect       unitFileState = "indirect"
	generated      unitFileState = "generated"
	transient      unitFileState = "transient"
	bad            unitFileState = "bad"
)

func allUnitFileStateStrings() []string {
	return []string{
		string(enabled),
		string(enabledRuntime),
		string(linked),
		string(linkedRuntime),
		string(alias),
		string(masked),
		string(maskedRuntime),
		string(static),
		string(disabled),
		string(indirect),
		string(generated),
		string(transient),
		string(bad),
	}
}
