<script lang="ts">
	import Navbar from "$lib/Navbar.svelte";
	import Container from "$lib/Container.svelte";
	export let data; // This will contain the data fetched in +page.server.ts
	let commands = data.commands;
</script>

<Container>
	<Navbar />
	<div
		class="overflow-x-auto rounded-box border border-base-content/5 bg-base-100"
	>
		<table class="table">
			<thead>
				<tr>
					<th>Name</th>
					<th>Usage</th>
					<th>Description</th>
				</tr>
			</thead>
			<tbody>
				{#each commands as cmd}
					<tr>
						<td>{cmd.Name}</td>
						<td>{cmd.Usage}</td>
						<td>{cmd.Description}</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
</Container>
