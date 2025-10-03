<template>
	<div :id='container.id' :class='"container-status " + container.state'>
		<h5>{{ container.names[0].replace(/^\/*/, "") }}</h5>
		<p>{{ container.state == "running" ? container.status?.toLocaleLowerCase() ?? container.state : container.state }}</p>
	</div>
</template>
<script setup lang='ts'>
import type { Container } from '@/types/container';

const container = defineProps<Container>();
</script>
<style scoped>
.container-status {
	display: block;
	margin: var(--half-margin);
	color: var(--dark);
	border-radius: var(--border-radius);
	width: 8em;
	overflow: hidden;
	text-wrap: nowrap;
	padding: 0.5em;
}

h5 {
	margin: 0;
}
p {
	margin: 0;
	font-size: 0.8em;
}

.running {
	background: var(--green);
}

.restarting {
	background: var(--warning);
}

.dead {
	background: var(--error);
}

.created, .paused, .exited, .removing {
	background: var(--bg1);
	color: var(--text);
}
</style>
