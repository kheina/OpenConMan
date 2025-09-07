<template>
	<div
		ref='div'
		class='code-editor'
		@keydown.tab.prevent='tab'
	/>
</template>
<script setup lang='ts'>
import { onMounted, ref, type Ref } from 'vue';
import { EditorView, basicSetup } from 'codemirror';
import { tags as t } from '@lezer/highlight';
// import { go } from '@codemirror/lang-go';
import { HighlightStyle, syntaxHighlighting } from '@codemirror/language';
import { ViewPlugin, ViewUpdate } from '@codemirror/view';

const div = ref<HTMLDivElement | null>(null) as Ref<HTMLDivElement>;
const props = defineProps<{ value?: string }>();
const emits = defineEmits(["update:value", "change", "input"]);
const updater = ViewPlugin.fromClass(class {
	update(update: ViewUpdate) {
		if (!update.docChanged) return;
		emits("update:value", update.state.doc.toString());
	}
});

onMounted(() => {
	const vscode = HighlightStyle.define([
		// this style is copied wholesale from https://github.com/uiwjs/react-codemirror/blob/cc35e8b0ca5f2eb18726fa6d0a3c4b0aa73841d1/themes/vscode/src/dark.ts#L20-L68 and only adjusted for code styling
		{
			tag: [
				t.keyword,
				t.operatorKeyword,
				t.modifier,
				t.color,
				t.constant(t.name),
				t.standard(t.name),
				t.standard(t.tagName),
				t.special(t.brace),
				t.atom,
				t.bool,
				t.special(t.variableName),
			],
			color: "#569cd6",
		},
		{ tag: [t.controlKeyword, t.moduleKeyword], color: "#c586c0" },
		{
			tag: [
				t.name,
				t.deleted,
				t.character,
				t.macroName,
				t.propertyName,
				t.variableName,
				t.labelName,
				t.definition(t.name),
			],
			color: "#9cdcfe",
		},
		{ tag: t.heading, fontWeight: "bold", color: "#9cdcfe" },
		{
			tag: [t.typeName, t.className, t.tagName, t.number, t.changed, t.annotation, t.self, t.namespace],
			color: "#4ec9b0",
		},
		{ tag: [t.function(t.variableName), t.function(t.propertyName)], color: "#dcdcaa" },
		{ tag: [t.number], color: "#b5cea8" },
		{ tag: [t.operator, t.punctuation, t.separator, t.url, t.escape, t.regexp], color: "#d4d4d4" },
		{ tag: [t.regexp], color: "#d16969" },
		{ tag: [t.special(t.string), t.processingInstruction, t.string, t.inserted], color: "#ce9178" },
		{ tag: [t.angleBracket], color: "#808080" },
		{ tag: t.strong, fontWeight: "bold" },
		{ tag: t.emphasis, fontStyle: "italic" },
		{ tag: t.strikethrough, textDecoration: "line-through" },
		{ tag: [t.meta, t.comment], color: "#6a9955" },
		{ tag: t.link, color: "#6a9955", textDecoration: "underline" },
		{ tag: t.invalid, color: "#ff0000" },
	]);
	const view = new EditorView({
		extensions: [
			syntaxHighlighting(vscode),
			EditorView.lineWrapping,
			basicSetup,
			updater,
		],
		parent: div.value,
		doc: props.value ?? "",
	});
});

function tab(e: KeyboardEvent) {
	console.debug("==> tab:", e);
}
</script>
<style scoped>
.code-editor :deep() {
	font-family: Hack, DejaVu Sans Mono, Inconsolata, monospace;
	border-radius: var(--border-radius);
	outline: none;

	.cm-content {
		overflow: hidden;
	}
	.cm-cursor {
		border-color: var(--text);
	}
	.cm-focused {
		outline: none;
	}
	.cm-line {
		white-space: pre-wrap;
	}
	.cm-activeLine {
		background: none;
		.cm-focused:not(:has(.cm-selectionBackground)) & {
			box-shadow: var(--border-color) 0 0 0 0.15em;
		}
	}
	.cm-gutters {
		background: none;
		border-color: var(--border-color);
		color: var(--subtle);
	}
	.cm-activeLineGutter {
		background: none;
	}
	.cm-selectionMatch {
		background: var(--border-color);
	}
	.cm-selectionLayer .cm-selectionBackground {
		background: var(--border-color);
		.cm-focused > .cm-scroller & {
			background: #264f78;
		}
	}
	/* .cm-selectionLayer {
		& :first-child {
			border-radius: var(--border-radius) var(--border-radius) 0 0;
		}
		not sure how to do middle children
		& :last-child {
			border-radius:0 0  var(--border-radius) var(--border-radius);
		}
	} */
	.cm-foldPlaceholder {
		background: none;
		border: none;
		color: var(--subtle);
		.cm-line:has(&) {
			background: #2f4a7466;
		}
	}
	.cm-tooltip {
		background: var(--bg3);
		color: var(--text);
		border-color: var(--border-color);
		border-radius: var(--border-radius);
		overflow: hidden;
	}
}
</style>
