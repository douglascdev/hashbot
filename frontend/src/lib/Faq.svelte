<script>
	export let faqs = [];
</script>

<div class="max-w-2xl mx-auto border border-base-300 rounded-lg p-4">
		<h1 class="text-2xl font-bold mb-4">FAQ</h1>
		{#each faqs as faq}
			<div class="collapse collapse-arrow bg-base-100 border border-base-300 mb-2">
				<input type="radio" name="my-accordion-2" />
				<div class="collapse-title font-semibold">
					{faq.title}
				</div>
				<div class="collapse-content text-sm">
					{@html faq.content}
				</div>
			</div>
		{/each}
</div>
