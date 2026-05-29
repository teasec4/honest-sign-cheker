<script lang="ts">
	import type { Problem } from './api';

	export let title: string;
	export let problems: Problem[] = [];
</script>

{#if problems.length > 0}
	<h3>{title}</h3>
	<table>
		<thead>
			<tr>
				<th>Тип</th>
				<th>Код</th>
				<th>Совпадение</th>
				<th>Кандидат</th>
				<th>Где</th>
			</tr>
		</thead>
		<tbody>
			{#each problems as problem}
				<tr>
					<td>{problem.type}</td>
					<td><code>{problem.code}</code></td>
					<td>{problem.matchPercent ? `${problem.matchPercent.toFixed(2)}%` : ''}</td>
					<td>
						{#if problem.matchedCode}
							<code>{problem.matchedCode}</code>
						{/if}
					</td>
					<td>
						{#if problem.returnedLocations?.length}
							<div>Возврат: {problem.returnedLocations.join(', ')}</div>
						{/if}
						{#if problem.issuedLocations?.length}
							<div>Выдача: {problem.issuedLocations.join(', ')}</div>
						{/if}
					</td>
				</tr>
			{/each}
		</tbody>
	</table>
{/if}
