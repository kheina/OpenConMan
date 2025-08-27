export interface Container {
	id: string,
	image: string,
	names: string[],
	state: "created" | "running" | "paused" | "restarting" | "exited" | "removing" | "dead",
	status: string,
}
