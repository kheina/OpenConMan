<template>
	<div class='centered'>
		<button @click='Notify'>sound</button>
	</div>
</template>
<script setup lang='ts'>
import { Notify } from '@/utilities';
import { ref, watch, type Ref } from 'vue';

const content: Ref<string> = ref("try me");
watch(content, (s: string) => console.log(s));
</script>
<style scoped>
.centered {
	display: flex;
	align-items: center;
	justify-content: center;
	height: 100%;
}
</style>
