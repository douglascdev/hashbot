<script>
	import Navbar from "$lib/Navbar.svelte";
	import Faq from "$lib/Faq.svelte";
	import Container from "$lib/Container.svelte";
	import Blog from "$lib/Blog.svelte";
	import { m } from "$lib/paraglide/messages";

	const blogPosts = [
		{
			title: "Hashbot",
			content: `
	<p>
		Hashbot is a <a
			href="https://dev.twitch.tv/docs/chat/"
			class="link link-primary">twitch chatbot</a
		>, still early in development, with functionalities such as:
	</p>
	<p>
		- ✅ 7TV emote management: lets you copy, remove, and add emotes,
		and manage your emote editors.
	</p>
	<p>
		- ✅ multi language support(portuguese and english)
    </p>
	<p>
		- ✅ buttsbot: randomly turn sentences into a buttified or
		glorpified version
	</p>
	<p>- ✅ enabling, disabling or opting out of commands.</p>
`,
		},
		{
			title: "Planned Features",
			content: `
<p>- ☐ <i>;song</i> command with Spotify integration.</p>
<p>- ☐ <i>;weather</i> command.</p>
<p>- ☐ <i>;setgame</i> command.</p>
`,
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
	<img alt="the bot waving at you" src="/glorp.png" class="center" />
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
