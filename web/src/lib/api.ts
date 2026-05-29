export interface FileSummary {
	total: number;
	unique: number;
}

export interface Problem {
	type: string;
	code: string;
	description: string;
	count?: number;
	matchPercent?: number;
	matchedCode?: string;
	returnedLocations?: string[];
	issuedLocations?: string[];
}

export interface PrimaryReport {
	minPercent: number;
	summary: {
		issued: FileSummary;
		returned: FileSummary;
		exactTotal: number;
		exactUnique: number;
		fuzzyTotal: number;
		fuzzyUnique: number;
		unknownTotal: number;
		unknownUnique: number;
		duplicateUnique: number;
	};
	duplicates: Problem[];
	fuzzy: Problem[];
	unknown: Problem[];
}

export interface DuplicateReport {
	summary: FileSummary;
	duplicates: Problem[];
}

const apiBase = import.meta.env.VITE_API_BASE ?? '';

async function postForm<T>(path: string, form: FormData): Promise<T> {
	const response = await fetch(`${apiBase}${path}`, {
		method: 'POST',
		body: form
	});
	const data = await response.json().catch(() => ({}));
	if (!response.ok) {
		throw new Error(data.error ?? `HTTP ${response.status}`);
	}
	return data as T;
}

export function primaryCheck(issued: File, returned: File, minPercent: number) {
	const form = new FormData();
	form.set('issued', issued);
	form.set('returned', returned);
	form.set('minPercent', String(minPercent));
	return postForm<PrimaryReport>('/api/primary-check', form);
}

export function duplicateCheck(restored: File) {
	const form = new FormData();
	form.set('restored', restored);
	return postForm<DuplicateReport>('/api/duplicate-check', form);
}
