<template>
	{{ unit }}
</template>
<script setup lang='ts'>
import { onUnmounted, ref, type Ref } from 'vue';
import type { UnitStatus } from '@/types/systemd';
import { cetch } from '@/utilities';

const unit: Ref<UnitStatus | undefined> = ref();
const props = defineProps<{
	name: string,
}>();

const abort = new AbortController();
onUnmounted(() => abort.abort());

cetch(`/v1/service/${props.name}?live=1`, {
	signal: abort.signal,
}).then(res => {
	if (!res.body) return;
	const ro = res.body.pipeThrough(
		new TextDecoderStream("utf-8", { "fatal": false }),
		{ signal: abort.signal },
	).getReader();

	const rf = () => {
		ro.read().then(r => {
			if (r.done) return;
			const rj: { item: UnitStatus } = JSON.parse(r.value).result;
			unit.value = rj.item
			rf();
		});
	};
	rf();
});
</script>
<style lang='css' scoped>
</style>
