<div class="navbar bg-base-100 shadow-sm">
	<div class="navbar-start">
		<div class="dropdown z-10">
			<div
				tabindex="0"
				role="button"
				class="btn btn-ghost lg:hidden"
			>
				<svg
					xmlns="http://www.w3.org/2000/svg"
					class="h-5 w-5"
					fill="none"
					viewBox="0 0 24 24"
					stroke="currentColor"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M4 6h16M4 12h8m-8 6h16"
					/>
				</svg>
			</div>
			<ul
				tabindex="0"
				class="menu menu-sm dropdown-content bg-base-100 rounded-box z-1 mt-3 w-52 p-2 shadow"
			>
				<li><a>Item 1</a></li>
				<li>
					<a>Parent</a>
					<ul class="p-2">
						<li><a>Submenu 1</a></li>
						<li><a>Submenu 2</a></li>
					</ul>
				</li>
				<li><a>Item 3</a></li>
			</ul>
		</div>
		<a class="btn btn-ghost text-xl" href="/">hashbot</a>
	</div>
	<div class="navbar-center hidden lg:flex">
		<ul class="menu menu-horizontal px-1">
			<li><a>Item 1</a></li>
			<li>
				<details>
					<summary>Commands</summary>
					<ul class="p-2 z-10">
						<li>
							<a href="/commands"
								>Full list of
								commands</a
							>
						</li>
						<li>
							<a
								href="/emote-management"
								>Emote
								management</a
							>
						</li>
						<li>
							<a href="/buttsbot"
								>Buttsbot</a
							>
						</li>
					</ul>
				</details>
			</li>
			<li><a>Item 3</a></li>
		</ul>
	</div>
	<div class="navbar-end"></div>
</div>
