<template>
	<div class='services' v-if='units'>
		<header>
			<div class='contents'>
				<button @click='() => {}'>
					new
				</button>
			</div>
		</header>
		<main>
			<div class='service' v-for='u in units.items'>
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
import { onMounted, onUnmounted, ref, type Ref } from 'vue';

const stopped: Set<string> = new Set(["dead", "disabled"]);
const units: Ref<{
	items: UnitStatus[],
} | null> = ref(null);
const update: Ref<number | undefined> = ref();

onMounted(() => update.value = Updater());
onUnmounted(() => update.value = clearTimeout(update.value) ?? undefined);

function StartService(u: UnitStatus) {
	fetch(
		`http://127.0.0.1:5050/v1/service/enable/${u.name}`
	).then(
		r => r.json()
	).then((r: { item: UnitStatus }) => {
		if (!units.value) return;
		const unit = units.value.items.findIndex((v: UnitStatus) => v.name === u.name);
		if (unit < 0) return;
		units.value.items[unit] = r.item;
	}).catch(
		console.error
	);
}

function StopService(u: UnitStatus) {
	fetch(
		`http://127.0.0.1:5050/v1/service/disable/${u.name}`
	).then(
		r => r.json()
	).then((r: { item: UnitStatus }) => {
		if (!units.value) return;
		const unit = units.value.items.findIndex((v: UnitStatus) => v.name === u.name);
		if (unit < 0) return;
		units.value.items[unit] = r.item;
	}).catch(
		console.error
	);
}


function Updater(): number {
	fetch(
		"http://127.0.0.1:5050/v1/services"
	).then(
		r => r.json()
	).then((r:{ items: UnitStatus[] }) =>
		units.value = r
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
	height: 100%;
	width: 100%;
	display: grid;
	grid-template-rows: [header-start] 3em [header-end] 0 [main-start] 1fr [main-end];
	grid-template-columns: [header-start main-start] auto [header-end main-end];
}
header {
	background: var(--bg2);
	grid-area: header;

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
	grid-area: main;
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
}
.indicator > div {
	width: 0.5em;
	height: 0.5em;
	border-radius: 50%;
	margin-right: 0.5em;
	background: var(--warning);

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
button {
	color: inherit;
	cursor: pointer;
	border-radius: var(--border-radius);
	padding: 0.5em 1em;
	border: var(--border-size) solid var(--border-color);
	display: inline-block;
	font-weight: normal;
	background: var(--bg2);
	box-sizing: border-box;
	box-shadow: 0 2px 3px 1px var(--shadowcolor);
	-webkit-transition: var(--transition) var(--fadetime);
	-moz-transition: var(--transition) var(--fadetime);
	-o-transition: var(--transition) var(--fadetime);
	transition: var(--transition) var(--fadetime);
}
button:hover {
	border-color: var(--borderhover);
	color: var(--interact);
	box-shadow: 0 0 10px 3px var(--activeshadowcolor);
}
</style>
