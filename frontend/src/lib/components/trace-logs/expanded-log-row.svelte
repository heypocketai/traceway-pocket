<script lang="ts" module>
	export type ExpandedLogRecord = {
		body: string;
		traceId?: string;
		spanId?: string;
		scopeName?: string;
		scopeVersion?: string;
		resourceAttributes?: Record<string, string> | null;
		scopeAttributes?: Record<string, string> | null;
		logAttributes?: Record<string, string> | null;
	};
</script>

<script lang="ts">
	import * as Table from '$lib/components/ui/table';
	import { AttributesGrid, AttributesView } from '$lib/components/ui/attributes-grid';

	let {
		log,
		colspan
	}: {
		log: ExpandedLogRecord;
		colspan: number;
	} = $props();

	function mergedAttributes(): Record<string, string> {
		const out: Record<string, string> = {};
		const push = (m: Record<string, string> | null | undefined, prefix: string) => {
			if (!m) return;
			for (const [k, v] of Object.entries(m)) {
				out[`${prefix}${k}`] = v;
			}
		};
		push(log.resourceAttributes, 'resource.');
		push(log.scopeAttributes, 'scope.');
		push(log.logAttributes, '');
		return out;
	}

	function hasAttributes(): boolean {
		for (const _ in mergedAttributes()) return true;
		return false;
	}
</script>

<!--
	The container <tr>/<td> carry a fixed background with `!` so Table.Row's
	built-in hover:[&>td]:bg-muted/50 can't repaint the cell on mouseover.
-->
<Table.Row class="!bg-background">
	<Table.Cell {colspan} class="!bg-background px-4 py-3">
		<div class="space-y-3">
			<div class="grid grid-cols-1 gap-3 sm:grid-cols-2 md:grid-cols-3">
				<div class="sm:col-span-2 md:col-span-3">
					<AttributesView title="Body" value={log.body} />
				</div>
				{#if log.traceId}
					<AttributesView title="Trace ID" value={log.traceId} />
				{/if}
				{#if log.spanId}
					<AttributesView title="Span ID" value={log.spanId} />
				{/if}
				{#if log.scopeName}
					<AttributesView
						title="Scope"
						value={log.scopeName + (log.scopeVersion ? `@${log.scopeVersion}` : '')}
					/>
				{/if}
			</div>
			{#if hasAttributes()}
				<AttributesGrid attributes={mergedAttributes()} collapsedCount={6} />
			{/if}
		</div>
	</Table.Cell>
</Table.Row>
