<template>
	<div class='centered'>
		<CodeEditor class='code-editor' v-model:value='content'/>
	</div>
</template>
<script setup lang='ts'>
import CodeEditor from '@/components/CodeEditor.vue';
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
.code-editor {
	max-width: 30em;
	min-width: 10em;

	padding: var(--margin);
	background: var(--bg1);
}
</style>
