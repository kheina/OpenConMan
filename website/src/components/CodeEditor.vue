<template>
	<div ref='div' class='code-editor'>
	</div>
</template>
<script setup lang='ts'>
import { onMounted, ref, type Ref } from 'vue';
import { EditorView, basicSetup } from "codemirror";
import { javascript } from "@codemirror/lang-javascript";
import { oneDark } from '@codemirror/theme-one-dark';

const div = ref<HTMLDivElement | null>(null) as Ref<HTMLDivElement>;
onMounted(() => {
	console.log(basicSetup)
	const view = new EditorView({
		extensions: [
			oneDark,
			basicSetup,
		],
		parent: div.value,
	});
});
</script>
<style scoped>
.code-editor :deep() {
	font-family: Hack, DejaVu Sans Mono, Inconsolata, monospace;
	padding: var(--margin);
	background: var(--bg1);
	outline: none;
	border-radius: var(--border-radius);

	.cm-activeLine {
		background: none;
		/* border-top: 0.15em solid var(--border-color);
		border-bottom: 0.15em solid var(--border-color);
		margin: -0.15em 0; */
	}
	.cm-gutters {
		background: var(--bg1);
		border-color: var(--border-color);
		color: var(--subtle);
	}
	.cm-activeLineGutter {
		background: none;
	}
}
</style>
