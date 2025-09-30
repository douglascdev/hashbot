<script>
	import Navbar from "$lib/Navbar.svelte";
	import Blog from "$lib/Blog.svelte";
	import Faq from "$lib/Faq.svelte";
	import Container from "$lib/Container.svelte";

	const blogPosts = [
		{
			title: "Adding emotes",
			content: `
<br/>
To add a single emote:
<div class="mockup-code w-full">
  <pre
    data-prefix=";"><code>add emote</code></pre>
</div>

<br/>
To add a single emote to a specified channel:
<div class="mockup-code w-full">
  <pre
    data-prefix=";"><code>add emote #channel</code></pre>
</div>

<br/>
The command also allows for adding multiple emotes at once:
<br/>
<div class="mockup-code w-full">
  <pre
    data-prefix=";"><code>add emote1 emote2 emote3 #channel</code></pre>
</div>
`,
		},
		{
			title: "Removing emotes",
			content: `
<br/>
To remove a single emote:
<div class="mockup-code w-full">
  <pre
    data-prefix=";"><code>remove emote</code></pre>
</div>

<br/>
To remove a single emote from a specified channel:
<div class="mockup-code w-full">
  <pre
    data-prefix=";"><code>remove emote #channel</code></pre>
</div>

<br/>
To remove multiple emotes from a specified channel:
<div class="mockup-code w-full">
  <pre
    data-prefix=";"><code>remove emote1 emote2 emote3 #channel</code></pre>
</div>
`,
		},
	];
	const faqs = [
		{
			title: "How do I rename an emote?",
			content: "Renaming isn't implemented yet",
		},
		{
			title: "How do I force add an emote(replace an existing emote)?",
			content: "There's no parameter to force add as of right now, so you must remove it then add the new one.",
		},
	];
</script>

<Container>
	<Navbar />
	<br />
	<Faq {faqs} />
	<br />

	<h1 class="text-2xl font-bold mb-4">Emote management</h1>
	<p>
		Hashbot provides several commands to help you manage your <a
			href="https://7tv.app/"
			class="link link-primary">7TV emotes</a
		>. Make sure to add the bot as a 7TV editor, otherwise you'll
		only get errors.
	</p>
	<Blog posts={blogPosts} />
</Container>
