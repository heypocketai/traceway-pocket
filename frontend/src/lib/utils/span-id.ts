export function spanIdUuidToHex(spanUuid: string | null | undefined): string {
	if (!spanUuid) return '';
	return spanUuid.replace(/-/g, '').slice(-16).toLowerCase();
}

export function traceIdUuidToHex(traceUuid: string | null | undefined): string {
	if (!traceUuid) return '';
	return traceUuid.replace(/-/g, '').toLowerCase();
}
