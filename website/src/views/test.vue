<template>
	<div class='centered'>
		<CodeEditor class='code-editor'/>
	</div>
</template>
<script setup lang='ts'>
import CodeEditor from '@/components/CodeEditor.vue';
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
}
</style>
