<template>
	<div class='dashboard'>
		<h2>Containers</h2>
		<div class='containers'>
			<ContainerStatus v-if='containers' v-for='c in containers.items' v-bind='c'/>
		</div>
		<h2>Networks</h2>
	</div>
</template>
<script setup lang='ts'>
import ContainerStatus from '@/components/ContainerStatus.vue';
import type { Container } from '@/types/container';
import { cetch } from '@/utilities';
import { ref, type Ref } from 'vue';

const containers: Ref<{
	items: Container[],
} | null> = ref(null);

cetch(
	"/v1/containers"
).then(
	r => r.json()
).then(r => {
	containers.value = r;
}).catch(
	console.error
)
</script>
<style scoped>
.dashboard {
	padding: var(--margin);
}
.dashboard > :first-child {
	margin-top: 0;
}

h2 {
	padding: 0 0 0 var(--margin);
	margin: var(--margin) var(--neg-half-margin) 0.5rem;
	border-bottom: var(--border-size) solid var(--border-color);
}

.containers {
	display: flex;
	flex-wrap: wrap;
	flex-direction: row;
	justify-content: flex-start;
	margin: var(--neg-half-margin);
}
</style>
