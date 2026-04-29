<script lang="ts">
	import * as Table from '$lib/components/ui/table';
	import { TracewayTableHeader } from '$lib/components/ui/traceway-table-header';
	import { TableEmptyState } from '$lib/components/ui/table-empty-state';
	import type { SessionLogEvent, SessionLogLevel } from '$lib/types/exceptions';

	interface Props {
		logs: SessionLogEvent[];
		startedAt?: string;
	}

	let { logs, startedAt }: Props = $props();

	function formatOffset(timestamp: string): string {
		if (!startedAt) return timestamp;
		const start = Date.parse(startedAt);
		const t = Date.parse(timestamp);
		if (!Number.isFinite(start) || !Number.isFinite(t)) return timestamp;
		const deltaMs = t - start;
		if (Math.abs(deltaMs) >= 1000) {
			const sign = deltaMs >= 0 ? '+' : '-';
			return `${sign}${(Math.abs(deltaMs) / 1000).toFixed(2)}s`;
		}
		const sign = deltaMs >= 0 ? '+' : '-';
		return `${sign}${Math.abs(deltaMs)}ms`;
	}

	function levelColor(level: SessionLogLevel): string {
		switch (level) {
			case 'error':
				return 'text-destructive';
			case 'warn':
				return 'text-yellow-600 dark:text-yellow-500';
			case 'debug':
				return 'text-muted-foreground';
			default:
				return 'text-foreground';
		}
	}
</script>

<div class="rounded-md border overflow-hidden">
	<Table.Root>
		{#if logs.length > 0}
			<Table.Header>
				<Table.Row>
					<TracewayTableHeader
						label="Time"
						tooltip="Offset from the start of the session recording"
						class="w-[110px]"
					/>
					<TracewayTableHeader
						label="Level"
						tooltip="Console severity (debug / info / warn / error)"
						class="w-[80px]"
					/>
					<TracewayTableHeader label="Message" tooltip="The console line as captured" />
				</Table.Row>
			</Table.Header>
		{/if}
		<Table.Body>
			{#if logs.length === 0}
				<TableEmptyState colspan={3} message="No logs captured for this session." />
			{:else}
				{#each logs as entry}
					<Table.Row>
						<Table.Cell class="font-mono text-xs text-muted-foreground tabular-nums">
							{formatOffset(entry.timestamp)}
						</Table.Cell>
						<Table.Cell class="font-mono text-xs uppercase {levelColor(entry.level)}">
							{entry.level}
						</Table.Cell>
						<Table.Cell
							class="font-mono text-xs whitespace-pre-wrap break-all"
							title={entry.message}
						>
							{entry.message}
						</Table.Cell>
					</Table.Row>
				{/each}
			{/if}
		</Table.Body>
	</Table.Root>
</div>
