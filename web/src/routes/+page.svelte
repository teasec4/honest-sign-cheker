<script lang="ts">
	import { primaryCheck, type PrimaryReport } from '$lib/api';
	import ProblemTable from '$lib/ProblemTable.svelte';

	let issuedFile = $state<File | null>(null);
	let returnedFile = $state<File | null>(null);
	let minPercent = $state(85);
	let loading = $state(false);
	let error = $state('');
	let report = $state<PrimaryReport | null>(null);

	let canSubmit = $derived(Boolean(issuedFile && returnedFile));

	function setFile(event: Event, target: 'issued' | 'returned') {
		const input = event.currentTarget as HTMLInputElement;
		const file = input.files?.[0] ?? null;
		if (target === 'issued') issuedFile = file;
		if (target === 'returned') returnedFile = file;
	}

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		if (!issuedFile || !returnedFile) return;
		loading = true;
		error = '';
		report = null;
		try {
			report = await primaryCheck(issuedFile, returnedFile, minPercent);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Ошибка проверки';
		} finally {
			loading = false;
		}
	}
</script>

<div class="grid">
	<form class="panel" onsubmit={submit}>
		<h2>Первичная сверка</h2>
		<label>
			Выданные коды
			<input type="file" accept=".csv,.txt,.xlsx,.xlsm" onchange={(event) => setFile(event, 'issued')} />
		</label>
		<label>
			Возврат поставщика
			<input type="file" accept=".csv,.txt,.xlsx,.xlsm" onchange={(event) => setFile(event, 'returned')} />
		</label>
		<label>
			Порог похожести, %
			<input type="number" min="60" max="100" bind:value={minPercent} />
		</label>
		<button disabled={!canSubmit || loading}>{loading ? 'Проверяю...' : 'Проверить'}</button>
	</form>

	<section class="panel">
		{#if error}
			<div class="error">{error}</div>
		{/if}

		{#if report}
			<div class="summary">
				<div><span>Выдано</span><strong>{report.summary.issued.total}</strong><small>{report.summary.issued.unique} уник.</small></div>
				<div><span>Возврат</span><strong>{report.summary.returned.total}</strong><small>{report.summary.returned.unique} уник.</small></div>
				<div><span>Точно</span><strong>{report.summary.exactUnique}</strong><small>{report.summary.exactTotal} кодов</small></div>
				<div><span>Восстановить</span><strong>{report.summary.fuzzyUnique}</strong><small>{report.summary.fuzzyTotal} кодов</small></div>
				<div><span>Не найдены</span><strong>{report.summary.unknownUnique}</strong><small>{report.summary.unknownTotal} кодов</small></div>
				<div><span>Дубли</span><strong>{report.summary.duplicateUnique}</strong></div>
			</div>

			<ProblemTable title="Дубликаты в возврате" problems={report.duplicates} />
			<ProblemTable title="Кандидаты на восстановление" problems={report.fuzzy} />
			<ProblemTable title="Не распознаны" problems={report.unknown} />

			<h3>JSON</h3>
			<pre>{JSON.stringify(report, null, 2)}</pre>
		{:else}
			<p class="muted">Запусти проверку, результат появится здесь.</p>
		{/if}
	</section>
</div>
