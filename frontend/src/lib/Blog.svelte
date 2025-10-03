<script>
	export let posts = []; // Array of blog post objects
</script>

{#each posts as post}
	<article class="shadow-md rounded-lg p-6 mb-6">
		<h2 class="text-2xl font-semibold">{post.title}</h2>
		<p class="mt-2">{@html post.content}</p>
	</article>
{/each}
