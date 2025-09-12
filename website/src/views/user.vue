<template>
	<div class='centered'>
		<main>
			at least this one
		</main>
	</div>
</template>
<script setup lang='ts'>
import { cetch } from '@/utilities';
import { ref, type Ref } from 'vue';

const username: Ref<string> = ref("");
const password: Ref<string> = ref("");
const host = `${window.location.protocol}//${window.location.hostname}:5050`;

function sendLogin() {
	cetch(`${host}/v1/user/login`, {
		method: "POST",
		body: JSON.stringify({
			username: username.value,
			password: password.value,
		}),
	}).then(
		r => r.json()
	).then((r: { token: string }) => {
		console.log(r);
	}).catch(
		console.error
	);
}
</script>
<style scoped>
.centered {
	display: flex;
	align-items: center;
	justify-content: center;
	height: 100%;
}
</style>
