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
				</div>
			</div>
			<div class='editor' v-if='newUnitContent'>
				<div>
					<div>
						<span>{{ (!newUnitName || newUnitName.endsWith(".service")) ? newUnitName : newUnitName + ".service" }}</span>
						<i class='material-icons-round' @click='() => newUnitContent = undefined'>close</i>
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
		</main>
	</div>
	<div class='loading' v-else>
		loading
	</div>
</template>
<script setup lang='ts'>
import { onMounted, onUnmounted, ref, type Ref } from 'vue';
import type { UnitStatus } from '@/types/systemd'; 
import unitfiletemplate from '@/constants/unit_file';
import CodeEditor from '@/components/CodeEditor.vue';
import { cetch } from '@/utilities';

const host = `${window.location.protocol}//${window.location.hostname}:5050`;
const stopped: Set<string> = new Set(["dead", "disabled"]);
const units: Ref<UnitStatus[] | null> = ref(null);
const update: Ref<number | undefined> = ref();
const newUnitContent: Ref<string | undefined> = ref();
const newUnitName: Ref<string | undefined> = ref();

onMounted(() => update.value = Updater());
onUnmounted(() => update.value = clearTimeout(update.value) ?? undefined);

function StartService(u: UnitStatus) {
	cetch(
		`${host}/v1/service/enable/${u.name}`
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
		`${host}/v1/service/disable/${u.name}`
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
	cetch(`${host}/v1/service/alias/${u.alias}`, {
		method: "DELETE",
	}).catch(
		console.error
	);
}

function DeleteService(u: UnitStatus) {
	cetch(`${host}/v1/service/${u.name}`, {
		method: "DELETE",
	}).catch(
		console.error
	);
}

function CreateService() {
	cetch(`${host}/v1/service`, {
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

function Updater(): number {
	cetch(
		`${host}/v1/services`
	).then(
		r => r.json()
	).then((r:{ items: UnitStatus[] }) =>
		units.value = r.items
	).then(() => {
		if (update.value !== undefined) update.value = setTimeout(Updater, 1000);
	}).catch(
		console.error
	);
	return 0;
}

function UnitState(u: UnitStatus): string {
	return u.sub_state || u.active_state || u.load_state;
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
	flex-direction: column;
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
	margin: 0;
}
.service button {
	background: var(--bg2);
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

.editor {
	position: absolute;
	width: 100%;
	height: calc(100vh - 3em);
	display: flex;
	justify-content: center;
	align-items: center;
	top: 3em;

	&>div {
		background: var(--bg3);
		border-radius: var(--border-radius);
		display: flex;
		flex-direction: column;
		box-shadow: 0 2px 3px 1px var(--shadowcolor);
		max-height: calc(100% - var(--margin) * 2);
		max-width: calc(100% - var(--margin) * 2);

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
			max-height: 80vh;
			max-width: 100%;
			overflow: scroll;
		}
	}
}
</style>
