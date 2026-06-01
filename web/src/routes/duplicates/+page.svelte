<script lang="ts">
	import { duplicateCheck, type DuplicateReport } from '$lib/api';
	import ProblemTable from '$lib/ProblemTable.svelte';

	let restoredFile = $state<File | null>(null);
	let loading = $state(false);
	let error = $state('');
	let report = $state<DuplicateReport | null>(null);

	let canSubmit = $derived(Boolean(restoredFile));

	function setFile(event: Event) {
		const input = event.currentTarget as HTMLInputElement;
		restoredFile = input.files?.[0] ?? null;
	}

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		if (!restoredFile) return;
		loading = true;
		error = '';
		report = null;
		try {
			report = await duplicateCheck(restoredFile);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Ошибка проверки';
		} finally {
			loading = false;
		}
	}
</script>

<div class="grid">
	<form class="panel" onsubmit={submit}>
		<h2>Проверка дублей</h2>
		<label>
			Восстановленный файл
			<input type="file" accept=".csv,.txt,.xlsx,.xlsm" onchange={setFile} />
		</label>
		<button disabled={!canSubmit || loading}>{loading ? 'Проверяю...' : 'Проверить'}</button>
	</form>

	<section class="panel">
		{#if error}
			<div class="error">{error}</div>
		{/if}

		{#if report}
			<div class="summary">
				<div><span>Всего</span><strong>{report.summary.total}</strong></div>
				<div><span>Уникальных</span><strong>{report.summary.unique}</strong></div>
				<div><span>Дубли</span><strong>{report.duplicates.length}</strong></div>
			</div>

			<ProblemTable title="Дубликаты" problems={report.duplicates} />

			<h3>JSON</h3>
			<pre>{JSON.stringify(report, null, 2)}</pre>
		{:else}
			<p class="muted">Запусти проверку, результат появится здесь.</p>
		{/if}
	</section>
</div>
