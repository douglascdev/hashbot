<script>
	import Navbar from "$lib/Navbar.svelte";
	import Faq from "$lib/Faq.svelte";
	import Container from "$lib/Container.svelte";
	import Blog from "$lib/Blog.svelte";
	import { m } from "$lib/paraglide/messages";

	const blogPosts = [
		{
			title: "Hashbot",
			content: m.main_page_content(),
		},
		{
			title: m.main_page_planned_features_title(),
			content: m.main_page_planned_features_content(),
		},
	];
	const faqs = [
		{
			title: m.faq_add_bot_to_channel(),
			content: m.faq_add_bot_to_channel_content(),
		},
		{
			title: m.faq_help(),
			content: m.faq_help_content(),
		},
		{
			title: m.faq_use_command(),
			content: m.faq_use_command_content(),
		},
	];
</script>

<Container>
	<Navbar />

	<br />
	<img
		alt={m.main_page_glorp_description()}
		src="/glorp.png"
		class="center"
	/>
	<br />
	<Faq {faqs} />

	<Blog posts={blogPosts} />
</Container>

<style>
	img.center {
		display: block;
		margin: 0 auto;
	}
</style>
