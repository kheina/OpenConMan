<template>
	<main class='containers'>
		<div class='container' v-if='containers' v-for='c in containers.items'>
			<div class='indicator'>
				<div :class='c.state'/>
				<h2>{{ Name(c) }}</h2>
			</div>
			<p>{{ c.image }}</p>
			<p>{{ c.state == "running" ? c.status.toLocaleLowerCase() : c.state }}</p>
		</div>
		<div class='loading' v-else>
			loading
		</div>
	</main>
</template>
<script setup lang='ts'>
import type { Container } from '@/types/container';
import { ref, type Ref } from 'vue';

const containers: Ref<{
	items: Container[],
} | null> = ref(null);

fetch(
	"https://127.0.0.1:5050/v1/containers"
).then(
	r => r.json()
).then(r =>
	containers.value = r
).catch(
	console.error
)

function Name(c: Container): string {
	if (c.names.length > 0) {
		return c.names[0].replace(/^\/*/, "");
	}
	return "unknown";
}
</script>
<style scoped>
h2, p {
	margin: 0;
}
main {
	height: 100vh;
	max-height: 100vh;
	position: absolute;
	width: 100%;
}
.container {
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

	&.running {
		background: var(--green);
	}
	&.restarting {
		background: var(--warning);
	}
	&.dead {
		background: var(--error);
	}
	&.created, &.paused, &.exited, &.removing {
		background: var(--bg2);
	}
}
</style>
