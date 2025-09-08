<template>
	<div class='services' v-if='units'>
		<header>
			<div class='contents'>
				<input v-model='search' placeholder='filter'>
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
					<button @click='() => CreateAlias(u)'>
						create ocm alias
					</button>
				</div>
			</div>
		</main>
	</div>
	<div class='loading' v-else>
		loading
	</div>
</template>
<script setup lang='ts'>
import type { UnitStatus } from '@/types/systemd'; 
import { ref, watch, type Ref } from 'vue';
import fuzzysort from 'fuzzysort';

interface PreparedStatus extends UnitStatus {
	_prepared: Fuzzysort.Prepared,
}

const host = `${window.location.protocol}//${window.location.hostname}:5050`;
const search: Ref<string | void> = ref();
const allUnits: PreparedStatus[] = [];
const units: Ref<PreparedStatus[] | null> = ref(null);

function CreateAlias(u: UnitStatus) {
	fetch(`${host}/v1/service/alias`, {
		method: "PUT",
		body: JSON.stringify({
			name: u.name,
		}),
	}).then(
		r => r.json()
	).then((r: { item: UnitStatus }) => {
		// if (!units.value) return;
		// const unit = units.value.items.findIndex((v: UnitStatus) => v.name === u.name);
		// if (unit < 0) return;
		// units.value.items[unit] = r.item;
	}).catch(
		console.error
	);
}

fetch(
	`${host}/v1/services/all`
).then(
	r => r.json()
).then((r:{ items: UnitStatus[] }) => {
	for (const item of r.items) {
		allUnits.push({
			...item,
			_prepared: fuzzysort.prepare(item.name),
		});
	}
	units.value = allUnits;
}).catch(
	console.error
);

function UnitState(u: UnitStatus): string {
	return u.sub_state || u.active_state || u.load_state;
}

let timeout: number | undefined;
const thr = 0.5;
const lim = 100;
watch(search, (value: string | void) => {
	clearTimeout(timeout);
	if (!value) units.value = allUnits;
	else {
		const u: { score: number, value: PreparedStatus }[] = [];
		timeout = setTimeout(() => {
			for (const unit of allUnits) {
				const result = fuzzysort.single(value, unit._prepared);
				if (result && result.score >= thr) {
					u.push({ score: result.score, value: unit });
					if (u.length >= lim) break;
				}
			}
			units.value = u.sort((a, b) => b.score - a.score).map(x => x.value);
		}, 250);
	}
});

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

	&.running, &.active, &.mounted, &.plugged, &.listening {
		background: var(--green);
	}
	&.restarting {
		background: var(--warning);
	}
	&.dead {
		background: var(--error);
	}
	&.disabled, &.exited {
		background: var(--dark);
	}
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
input:focus::placeholder {
	color: #eeeeee20;
}
input:hover {
	border-color: var(--borderhover);
	color: var(--interact);
	box-shadow: 0 0 10px 3px var(--activeshadowcolor);
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
	-webkit-transition: var(--transition) var(--fadetime);
	-moz-transition: var(--transition) var(--fadetime);
	-o-transition: var(--transition) var(--fadetime);
	transition: var(--transition) var(--fadetime);
}
.service button {
	background: var(--bg2);
}
button:hover {
	border-color: var(--borderhover);
	color: var(--interact);
	box-shadow: 0 0 10px 3px var(--activeshadowcolor);
}

</style>
