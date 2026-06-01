<script lang="ts">
	import '../app.css';
	import { page } from '$app/state';

	let { children } = $props();
</script>

<main>
	<h1>Проверка кодов</h1>
	<nav>
		<a href="/" aria-current={page.url.pathname === '/' ? 'page' : undefined}>Первичная сверка</a>
		<a href="/duplicates" aria-current={page.url.pathname === '/duplicates' ? 'page' : undefined}>Дубликаты</a>
	</nav>
	{@render children()}
</main>
