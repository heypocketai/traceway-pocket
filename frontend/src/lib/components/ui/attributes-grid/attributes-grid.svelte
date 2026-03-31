<script lang="ts">
	import AttributesView from './attributes-view.svelte';

	let {
		attributes,
		sorted = true,
		collapsedCount = 3
	}: {
		attributes: Record<string, string>;
		sorted?: boolean;
		collapsedCount?: number;
	} = $props();

	let expanded = $state(false);

	const entries = $derived(() => {
		const items = Object.entries(attributes);
		if (sorted) {
			return items.sort((a, b) => a[0].localeCompare(b[0]));
		}
		return items;
	});

	const visibleEntries = $derived(() => {
		const all = entries();
		if (expanded || all.length <= collapsedCount) return all;
		return all.slice(0, collapsedCount);
	});

	const hiddenCount = $derived(() => {
		const all = entries();
		if (expanded || all.length <= collapsedCount) return 0;
		return all.length - collapsedCount;
	});
</script>

<div>
	<div class="grid grid-cols-1 gap-3 sm:grid-cols-2 md:grid-cols-3">
		{#each visibleEntries() as [key, value]}
			<AttributesView title={key} {value} />
		{/each}
	</div>
	{#if hiddenCount() > 0}
		<div class="flex justify-end">
			<button
				onclick={() => expanded = true}
				class="mt-2 text-xs text-muted-foreground hover:text-foreground transition-colors"
			>
				Show {hiddenCount()} more...
			</button>
		</div>
	{:else if expanded && entries().length > collapsedCount}
		<div class="flex justify-end">
			<button
				onclick={() => expanded = false}
				class="mt-2 text-xs text-muted-foreground hover:text-foreground transition-colors"
			>
				Show less
			</button>
		</div>
	{/if}
</div>
