type FormatResult = { text: string; unit: string };

export function formatMetricValue(value: number, unit: string): FormatResult {
	if (unit === '%') {
		if (value === 0) return { text: '0', unit: '%' };
		if (Math.abs(value) < 0.1) return { text: value.toFixed(2), unit: '%' };
		if (Math.abs(value) < 10) return { text: value.toFixed(1), unit: '%' };
		return { text: Math.round(value).toString(), unit: '%' };
	}

	if (unit === 'ms') {
		if (value < 1) return { text: (value * 1000).toFixed(0), unit: 'µs' };
		if (value < 10) return { text: value.toFixed(1), unit: 'ms' };
		if (value < 1000) return { text: Math.round(value).toString(), unit: 'ms' };
		return { text: (value / 1000).toFixed(1), unit: 's' };
	}

	if (unit === 'ns') {
		if (value < 1000) return { text: Math.round(value).toString(), unit: 'ns' };
		if (value < 1_000_000) return { text: (value / 1000).toFixed(1), unit: 'µs' };
		if (value < 1_000_000_000) return { text: (value / 1_000_000).toFixed(1), unit: 'ms' };
		return { text: (value / 1_000_000_000).toFixed(1), unit: 's' };
	}

	if (unit === 's') {
		if (value < 0.001) return { text: (value * 1_000_000).toFixed(0), unit: 'µs' };
		if (value < 1) return { text: (value * 1000).toFixed(1), unit: 'ms' };
		if (value < 60) return { text: value.toFixed(1), unit: 's' };
		if (value < 3600) return { text: (value / 60).toFixed(1), unit: 'min' };
		return { text: (value / 3600).toFixed(1), unit: 'h' };
	}

	if (unit === 'MB') {
		if (value < 1) return { text: (value * 1024).toFixed(0), unit: 'KB' };
		if (value >= 1024) return { text: (value / 1024).toFixed(1), unit: 'GB' };
		return { text: Math.round(value).toString(), unit: 'MB' };
	}

	if (unit === 'GB') {
		if (value < 1) return { text: (value * 1024).toFixed(0), unit: 'MB' };
		if (value >= 1024) return { text: (value / 1024).toFixed(1), unit: 'TB' };
		return { text: value.toFixed(1), unit: 'GB' };
	}

	if (unit === 'bytes' || unit === 'B') {
		if (value < 1024) return { text: Math.round(value).toString(), unit: 'B' };
		if (value < 1024 * 1024) return { text: (value / 1024).toFixed(1), unit: 'KB' };
		if (value < 1024 * 1024 * 1024) return { text: (value / (1024 * 1024)).toFixed(1), unit: 'MB' };
		return { text: (value / (1024 * 1024 * 1024)).toFixed(1), unit: 'GB' };
	}

	if (unit === 'count' || unit === '') {
		if (value >= 1_000_000) return { text: (value / 1_000_000).toFixed(1), unit: 'M' };
		if (value >= 1_000) return { text: (value / 1_000).toFixed(1), unit: 'K' };
		return { text: Math.round(value).toString(), unit: '' };
	}

	if (Number.isInteger(value)) return { text: value.toString(), unit };
	return { text: value.toFixed(1), unit };
}

export function formatMetricLabel(value: number, unit: string): string {
	const result = formatMetricValue(value, unit);
	if (!result.unit) return result.text;
	return `${result.text} ${result.unit}`;
}
