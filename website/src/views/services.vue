<template>
	<div class='services' v-if='units'>
		<header>
			<div class='contents'>
				<button @click='() => newUnitContent = unitfiletemplate'>
					new
				</button>
			</div>
		</header>
		<main>
			<div class='service' v-for='u in units'>
				<div class='data'>
					<div class='indicator'>
						<div :class='UnitState(u)'/>
						<h2>{{ u.name }}</h2>
					</div>
					<p>{{ u.description }}</p>
					<p>{{ u.sub_state }}</p>
				</div>
				<div class='buttons'>
					<button @click='() => StartService(u)' v-if='stopped.has(UnitState(u))'>
						enable
					</button>
					<button @click='() => StopService(u)' v-else>
						disable
					</button>
					<button @click='() => DeleteAlias(u)' v-if='u.alias'>
						delete
					</button>
					<button @click='() => DeleteService(u)' v-else>
						delete
					</button>
					<button @click='() => GetServiceLogs(u)'>
						logs
					</button>
				</div>
			</div>
			<div class='window'>
				<div class='editor' v-if='newUnitContent'>
					<div>
						<div>
							<span>{{ (!newUnitName || newUnitName.endsWith(".service")) ? newUnitName : newUnitName + ".service" }}</span>
							<i class='material-icons-round' @click='() => newUnitName = newUnitContent = undefined'>close</i>
						</div>
						<CodeEditor class='code-editor' v-model:value='newUnitContent'/>
						<div>
							<p>add <code>ocm-</code> to the start of the unit name for dashboard tracking</p>
							<div>
								<input v-model='newUnitName' placeholder='new-unit.service'/>
								<button @click='CreateService'>
									create unit
								</button>
							</div>
						</div>
					</div>
				</div>
				<div class='logs' v-else-if='logs !== undefined'>
					<div v-if='logs === null'>
						<div>
							<span>loading</span>
						</div>
					</div>
					<div v-else>
						<div>
							<span>logs: {{ logs.name }}</span>
							<i class='material-icons-round' @click='() => { abort(); logs = undefined; }'>close</i>
						</div>
						<div v-if='logs.logs'>
							<code v-for='l in logs.logs' v-html='LogMessage(l.fields)'/>
							<button  @click='() => GetServiceLogs(logs ?? { name: "" })' v-show='showmore'>load more</button>
						</div>
						<div v-else>
							<div>
								none
							</div>
						</div>
					</div>
				</div>
			</div>
		</main>
	</div>
	<div class='loading' v-else>
		loading
	</div>
</template>
<script setup lang='ts'>
import { onUnmounted, ref, type Ref } from 'vue';
import type { UnitLogs, UnitStatus } from '@/types/systemd'; 
import unitfiletemplate from '@/constants/unit_file';
import CodeEditor from '@/components/CodeEditor.vue';
import { cetch, JsonPipeThrough } from '@/utilities';

const stopped: Set<string> = new Set(["dead", "disabled"]);
const units: Ref<UnitStatus[] | null> = ref(null);
const newUnitContent: Ref<string | undefined> = ref();
const newUnitName: Ref<string | undefined> = ref();
const logs: Ref<UnitLogs | null | undefined> = ref(); // null == loading
const showmore: Ref<Boolean> = ref(false);

let _abort = new AbortController();
const abort = () => {
	_abort.abort();
	_abort = new AbortController();
};
onUnmounted(() => _abort.abort());
onUnmounted(() => svcAbort.abort());

const svcAbort = new AbortController();

cetch("/v1/services?live=1", {
	signal: svcAbort.signal,
}).then(res => JsonPipeThrough(res, svcAbort, (r: { items: UnitStatus[] }) => {
	units.value = r.items;
})).catch(
	console.error
);

function StartService(u: UnitStatus) {
	cetch(
		`/v1/service/enable/${u.name}`
	).then(
		r => r.json()
	).then((r: { item: UnitStatus }) => {
		if (!units.value) return;
		const unit = units.value.findIndex((v: UnitStatus) => v.name === u.name);
		if (unit < 0) return;
		units.value[unit] = r.item;
	}).catch(
		console.error
	);
}

function StopService(u: UnitStatus) {
	cetch(
		`/v1/service/disable/${u.name}`
	).then(
		r => r.json()
	).then((r: { item: UnitStatus }) => {
		if (!units.value) return;
		const unit = units.value.findIndex((v: UnitStatus) => v.name === u.name);
		if (unit < 0) return;
		units.value[unit] = r.item;
	}).catch(
		console.error
	);
}

function DeleteAlias(u: UnitStatus) {
	if (!u.alias) return;
	cetch(`/v1/service/alias/${u.alias}`, {
		method: "DELETE",
	}).catch(
		console.error
	);
}

function DeleteService(u: UnitStatus) {
	cetch(`/v1/service/${u.name}`, {
		method: "DELETE",
	}).catch(
		console.error
	);
}

function GetServiceLogs(u: { name: string }) {
	showmore.value = false;
	let url = `/v1/service/logs/${u.name}`
	const lines = 100;
	let live = false;
	if (logs.value) {
		url += "?seek=" + encodeURIComponent(logs.value.logs[logs.value.logs.length-1].cursor);
	} else {
		logs.value = null;
		url += "?live=1";
		live = true;
	}
	url += `&lines=${lines}`;

	const abort = _abort;
	cetch(url, {
		signal: abort.signal,
	}).then(res => JsonPipeThrough(res, abort, (r: UnitLogs) => {
		if (live) {
			if (logs.value) {
				logs.value.logs.unshift(...r.logs);
			} else {
				logs.value = r;
				showmore.value = r.logs.length === lines;
			}
		} else {
			if (logs.value) {
				logs.value.logs = logs.value.logs.concat(r.logs);
			} else {
				logs.value = r;
			}
			showmore.value = r.logs.length === lines;
		}
	})).catch(
		console.error
	);
}

function CreateService() {
	cetch("/v1/service", {
		method: "PUT",
		body: JSON.stringify({
			name: newUnitName.value,
			content: newUnitContent.value,
		}),
	}).then(
		r => r.json()
	).then((r: UnitStatus) => {
		newUnitName.value = undefined;
		newUnitContent.value = undefined;
	}).catch(
		console.error
	);
}

function UnitState(u: UnitStatus): string {
	return u.sub_state || u.active_state || u.load_state;
}

const re = /\u001b\[\d{1,2}(;\d{1,2})*m/g;
// function colorize(str: string): string {
// 	let fg = "";
// 	let bg = "";
// 	let st = "";
// 	const fore = (f: string): string => {
// 		let r = "";
// 		if (fg) r += "</span>";
// 		fg = f;
// 		return r + `<span class="${fg}">`
// 	};
// 	const back = (b: string): string => {
// 		let r = "";
// 		if (fg) r += "</span>";
// 		if (bg) r += "</span>";
// 		bg = b;
// 		return r + `<span class="${fg} ${bg}">`
// 	};
// 	const style = (s: string): string => {
// 		let r = "";
// 		if (fg) r += "</span>";
// 		if (bg) r += "</span>";
// 		if (st) r += "</span>";
// 		st = s;
// 		return r + `<span class="${fg} ${bg} ${st}">`
// 	};
// 	const reset = (): string => {
// 		let r = "";
// 		if (fg) {
// 			r += "</span>";
// 			fg = "";
// 		}
// 		if (bg) {
// 			r += "</span>";
// 			bg = "";
// 		}
// 		if (st) {
// 			r += "</span>";
// 			st = "";
// 		}
// 		return r;
// 	}

// 	str = str.replace(re, (_m: string) => {
// 		let r = "";
// 		// don't ask me why parse.Int doesn't work
// 		_m.substring(2, _m.length - 1).split(";").map(parseFloat).map(Math.round).forEach(m => {
// 			switch (m) {
// 			// STYLE
// 			// reset
// 			case 0:
// 				r += reset();
// 			// bold
// 			case 1:
// 				r += style("bold");
// 			// disable
// 			case 2:
// 				r += style("disable");
// 			// underline
// 			case 4:
// 				r += style("underline");
// 			// reverse
// 			case 7:
// 				r += style("reverse");
// 			// strikethrough
// 			case 9:
// 				r += style("strikethrough");
// 			// invisible
// 			case 8:
// 				r += style("invisible");

// 			// FG
// 			// black
// 			case 30:
// 				r += fore("black");
// 			// red
// 			case 31:
// 				r += fore("red");
// 			// green
// 			case 32:
// 				r += fore("green");
// 			// orange
// 			case 33:
// 				r += fore("orange");
// 			// blue
// 			case 34:
// 				r += fore("blue");
// 			// purple
// 			case 35:
// 				r += fore("purple");
// 			// cyan
// 			case 36:
// 				r += fore("cyan");
// 			// lightgrey
// 			case 37:
// 				r += fore("lightgrey");
// 			// darkgrey
// 			case 90:
// 				r += fore("darkgrey");
// 			// lightred
// 			case 91:
// 				r += fore("lightred");
// 			// lightgreen
// 			case 92:
// 				r += fore("lightgreen");
// 			// yellow
// 			case 93:
// 				r += fore("yellow");
// 			// lightblue
// 			case 94:
// 				r += fore("lightblue");
// 			// pink
// 			case 95:
// 				r += fore("pink");
// 			// lightcyan
// 			case 96:
// 				r += fore("lightcyan");

// 			// BG
// 			// black
// 			case 40:
// 				r += back("black");
// 			// red
// 			case 41:
// 				r += back("black");
// 			// green
// 			case 42:
// 				r += back("black");
// 			// orange
// 			case 43:
// 				r += back("black");
// 			// blue
// 			case 44:
// 				r += back("black");
// 			// purple
// 			case 45:
// 				r += back("black");
// 			// cyan
// 			case 46:
// 				r += back("black");
// 			// lightgrey
// 			case 47:
// 				r += back("black");

// 			default:
// 				r = _m;
// 			}
// 		});
// 		return r;
// 	});

// 	reset();
// 	return str;
// }

function decolorize(msg: string): string {
	return msg.replace(re, "");
}

function LogMessage(fields: { [k: string]: string }): string {
	let message = fields.message;

	// this is a bit of a weird case, the logs we receive are RAW, so we basically
	// want to delete any message data after the final carriage return, as it
	// appears in journalctl
	const cr = message.lastIndexOf("\r");
	if (cr > 0 && message.length > cr + 1) message = message.substring(cr + 1);
	return decolorize(message);
}
</script>
<style scoped>
h2, p {
	margin: 0;
}
.services {
	position: relative;
}
header {
	background: var(--bg2);
	position: fixed;
	width: calc(100% - 15em);
	top: 0;
	height: 3em;

	& .contents {
		display: flex;
		align-items: center;
		margin: 0 var(--half-margin);
		padding: 0 var(--margin);
		height: 100%;
		border-bottom: var(--border-size) solid var(--border-color);
	}
}
main {
	padding-top: 3em;
}
.service {
	display: flex;
	justify-content: space-between;
	flex-direction: row;
	margin: var(--margin);
	padding: var(--margin);
	background: var(--bg1);
	border-radius: var(--border-radius);
}
.loading {
	height: 100%;
	display: flex;
	align-items: center;
	justify-content: center;
}

.indicator {
	display: flex;
	align-items: center;

	h2 {
		white-space: wrap;
		word-wrap: anywhere;
	}
}
.indicator > div {
	width: 0.5em;
	height: 0.5em;
	border-radius: 50%;
	margin-right: 0.5em;
	background: var(--warning);
	flex-shrink: 0;

	&.running {
		background: var(--green);
	}
	&.restarting {
		background: var(--warning);
	}
	&.dead {
		background: var(--error);
	}
	&.disabled {
		background: var(--dark);
	}
}
.buttons {
	display: flex;
	flex-direction: row;
	align-items: flex-start;
	justify-content: flex-end;
	flex-wrap: wrap;
	flex-shrink: 1;
}
button {
	color: inherit;
	cursor: pointer;
	border-radius: var(--border-radius);
	padding: 0.5em 1em;
	border: var(--border-size) solid var(--border-color);
	display: inline-block;
	font-weight: normal;
	background: var(--bg1);
	box-sizing: border-box;
	box-shadow: 0 2px 3px 1px var(--shadowcolor);
	margin-bottom: 1em;
	-webkit-transition: var(--transition) var(--fadetime);
	-moz-transition: var(--transition) var(--fadetime);
	-o-transition: var(--transition) var(--fadetime);
	transition: var(--transition) var(--fadetime);
}
button:last-child {
	margin-bottom: 0;
}
.service button {
	background: var(--bg2);
	margin: 0 0 1em 1em;
}
button:hover {
	border-color: var(--borderhover);
	color: var(--interact);
	box-shadow: 0 0 10px 3px var(--activeshadowcolor);
}

input {
	color: inherit;
	cursor: text;
	border-radius: var(--border-radius);
	padding: 0.5em 1em;
	border: var(--border-size) solid var(--border-color);
	display: inline-block;
	font-weight: normal;
	background: var(--bg1);
	box-sizing: border-box;
	box-shadow: 0 2px 3px 1px var(--shadowcolor);
	outline: none;

	&, &::placeholder {
		-webkit-transition: var(--transition) var(--fadetime);
		-moz-transition: var(--transition) var(--fadetime);
		-o-transition: var(--transition) var(--fadetime);
		transition: var(--transition) var(--fadetime);
	}
}
input:hover::placeholder {
	color: var(--interact);
}
input:hover {
	border-color: var(--borderhover);
	color: var(--interact);
	box-shadow: 0 0 10px 3px var(--activeshadowcolor);
}
input:focus {
	color: var(--text);
	&::placeholder {
		color: #eeeeee20;
	}
}

.window {
	position: fixed;
	top: 0;
	right: 0;
	width: calc(100vw - 15em - var(--border-size));
	height: 100vh;

	&:not(:has(*)) {
		display: none;
	}
}

.editor {
	position: absolute;
	width: 100%;
	height: 100vh;
	display: flex;
	justify-content: center;
	align-items: center;
	top: 0;

	&>div {
		background: var(--bg3);
		border-radius: var(--border-radius);
		display: flex;
		flex-direction: column;
		box-shadow: 0 2px 3px 1px var(--shadowcolor);
		max-height: calc(100% - 2em);
		max-width: calc(100% - 2em);

		&>div:first-child {
			display: flex;
			color: var(--subtle);
			flex-direction: row;
			align-items: center;
			justify-content: space-between;

			&>i {
				cursor: pointer;
				-webkit-transition: var(--transition) var(--fadetime);
				-moz-transition: var(--transition) var(--fadetime);
				-o-transition: var(--transition) var(--fadetime);
				transition: var(--transition) var(--fadetime);

				&:hover {
					color: var(--red);
				}
			}

			&>span {
				margin-left: 0.3em;
			}
		}

		&>div:last-child {
			margin: 0.25em 0.5em 0.5em;

			&>div {
				display: flex;
				width: 100%;
			}
			p {
				margin-bottom: 0.5em;
				font-size: 0.8em;
			}
			code {
				font-size: 1.25em;
			}
			input {
				margin-right: var(--margin);
				width: 100%;
			}
			button {
				white-space: nowrap;
				word-wrap: none;
			}
		}

		&>.code-editor {
			border-radius: 0;
			background: var(--bg1);
			min-width: 40em;
			min-height: 10em;
			/* max-height: 80vh;
			max-width: 100%; */
			overflow: scroll;
		}
	}
}

.logs {
	position: absolute;
	width: 100%;
	height: 100vh;
	display: flex;
	justify-content: center;
	align-items: center;
	top: 0;

	&>div {
		background: var(--bg3);
		border-radius: var(--border-radius);
		display: flex;
		flex-direction: column;
		box-shadow: 0 2px 3px 1px var(--shadowcolor);
		max-height: calc(100% - 2em);
		max-width: calc(100% - 2em);
		overflow: hidden;

		&>div:first-child {
			display: flex;
			color: var(--subtle);
			flex-direction: row;
			align-items: center;
			justify-content: space-between;

			&>i {
				cursor: pointer;
				-webkit-transition: var(--transition) var(--fadetime);
				-moz-transition: var(--transition) var(--fadetime);
				-o-transition: var(--transition) var(--fadetime);
				transition: var(--transition) var(--fadetime);

				&:hover {
					color: var(--red);
				}
			}

			&>span {
				margin-left: 0.3em;
			}
		}

		&>div:last-child {
			border-radius: 0;
			background: var(--bg1);
			min-width: 40em;
			min-height: 10em;
			overflow: scroll;
			padding: 0.25em;
			display: flex;
			flex-direction: column-reverse;

			&>div {
				margin: auto;
			}

			code {
				white-space: preserve-spaces;
				text-wrap: wrap;
				line-height: 1.2em;
			}

			button {
				align-self: center;
				margin: 0.5em;
			}
		}
	}
}
</style>
