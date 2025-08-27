export interface UnitStatus {
	// The primary unit name as string
	name: string,
	// The human readable description string
	description: string,
	// The load state (i.e. whether the unit file has been loaded successfully)
	load_state: string,
	// The active state (i.e. whether the unit is currently started or not)
	active_state: string,
	// The sub state (a more fine-grained version of the active state that is
	// specific to the unit type, which the active state is not)
	sub_state: string,
	// A unit that is being followed in its state by this unit, if there is any,
	// otherwise the empty string.
	followed: string,
	// The unit object path
	path: string,
	// If there is a job queued for the job unit the numeric job id, 0 otherwise
	job_id: Number,
	// The job type as string
	job_type: string,
	// The job object path
	job_path: string,
}
