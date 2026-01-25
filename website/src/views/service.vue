<template>
	{{ unit }}
</template>
<script setup lang='ts'>
import { onUnmounted, ref, type Ref } from 'vue';
import type { UnitStatus } from '@/types/systemd';
import { cetch, JsonPipeThrough } from '@/utilities';

const unit: Ref<UnitStatus | undefined> = ref();
const props = defineProps<{
	name: string,
}>();

const abort = new AbortController();
onUnmounted(() => abort.abort());

cetch(`/v1/service/${props.name}?live=1`, {
	signal: abort.signal,
}).then(res => JsonPipeThrough(res, abort, (rj: { item: UnitStatus }) => {
	unit.value = rj.item;
}));
</script>
<style lang='css' scoped>
</style>
