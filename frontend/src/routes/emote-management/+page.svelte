<script>
	import { m } from "$lib/paraglide/messages";
	import Navbar from "$lib/Navbar.svelte";
	import Blog from "$lib/Blog.svelte";
	import Faq from "$lib/Faq.svelte";
	import Container from "$lib/Container.svelte";
	import Alert from "$lib/Alert.svelte";

	const blogPosts = [
		{
			title: m.blog_add_emotes(),
			content: m.blog_add_emotes_content(),
		},
		{
			title: m.blog_remove_emotes(),
			content: m.blog_remove_emotes_content(),
		},
	];
	const faqs = [
		{
			title: m.faq_title_emote_commands(),
			content: m.faq_content_emote_commands(),
		},
		{
			title: m.faq_title_rename_emote(),
			content: m.faq_content_rename_emote(),
		},
		{
			title: m.faq_title_force_add(),
			content: m.faq_content_force_add(),
		},
	];
</script>

<Container>
	<Navbar />

	<br />
	<Alert alert={m.alert_editor_request()} />
	<br />

	<h1 class="text-2xl font-bold mb-4">{m.emote_management()}</h1>

	{@html m.index_content()}
	<Blog posts={blogPosts} />

	<br />
	<Faq {faqs} />
	<br />
</Container>
