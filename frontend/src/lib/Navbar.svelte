<script>
	import { m } from "./paraglide/messages";
</script>

<div class="navbar bg-base-100 shadow-sm">
	<div class="navbar-start">
		<a class="btn btn-ghost text-xl" href="/">hashbot</a>
	</div>
	<div class="navbar-center lg:flex">
		<ul class="menu menu-horizontal px-1">
			<li>
				<details class="max-w-1xl">
					<summary>{m.navbar_commands()}</summary>
					<ul
						class="p-2 z-10"
						style="min-width: max-content;"
					>
						<li>
							<a href="/commands"
								>{m.navbar_command_list()}</a
							>
						</li>
						<li>
							<a
								href="/emote-management"
								>{m.navbar_emote_management()}</a
							>
						</li>
					</ul>
				</details>
			</li>
		</ul>
	</div>
	<div class="navbar-end"></div>
</div>
