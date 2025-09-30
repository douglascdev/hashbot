<script>
	import { onMount } from "svelte";
	import Navbar from "$lib/Navbar.svelte";
	import Container from "$lib/Container.svelte";
	let commands = [];
	//	let data;

	onMount(async () => {
		//		const accessToken = getCookie("accessToken");
		//		fetch("https://id.twitch.tv/oauth2/validate", {
		//			method: "GET",
		//			headers: {
		//				Authorization: `OAuth ${accessToken}`,
		//			},
		//		})
		//			.then((response) => {
		//				if (!response.ok) {
		//					eraseCookie("accessToken");
		//					eraseCookie("refreshToken");
		//				}
		//			})
		//			.catch((error) => {
		//				console.error(
		//					"There was a problem with the fetch operation:",
		//					error,
		//				);
		//			});
		//
		//		let login = document.getElementById("login");
		//		let logout = document.getElementById("logout");
		//		if (isLoggedIn()) {
		//			login.setAttribute("hidden", "hidden");
		//			logout.removeAttribute("hidden");
		//		} else {
		//			logout.setAttribute("hidden", "hidden");
		//			login.removeAttribute("hidden");
		//		}
		const res = await fetch(
			`${import.meta.env.BASE_URL}commands.json`,
		);
		//		const dataReq = await fetch(
		//			`${import.meta.env.BASE_URL}client_data`,
		//		);
		commands = await res.json();
		//		data = await dataReq.json();
	});
	//
	//	function setCookie(name, value, days) {
	//		var expires = "";
	//		if (days) {
	//			var date = new Date();
	//			date.setTime(
	//				date.getTime() + days * 24 * 60 * 60 * 1000,
	//			);
	//			expires = "; expires=" + date.toUTCString();
	//		}
	//		document.cookie =
	//			name + "=" + (value || "") + expires + "; path=/";
	//	}
	//	function getCookie(name) {
	//		var nameEQ = name + "=";
	//		var ca = document.cookie.split(";");
	//		for (var i = 0; i < ca.length; i++) {
	//			var c = ca[i];
	//			while (c.charAt(0) == " ") c = c.substring(1, c.length);
	//			if (c.indexOf(nameEQ) == 0)
	//				return c.substring(nameEQ.length, c.length);
	//		}
	//		return null;
	//	}
	//	function eraseCookie(name) {
	//		document.cookie = name + "=; Max-Age=-99999999;";
	//	}
	//
	//	function twitchRedirect() {
	//		let state = Math.random().toString(36).slice(2);
	//		let redirectUri = data["redirect_uri"];
	//		let clientId = data["client_id"];
	//		let scopes = "channel:bot chat:edit chat:read";
	//		setCookie("loginState", state);
	//		window.location.href = `https://id.twitch.tv/oauth2/authorize?response_type=code&client_id=${clientId}&redirect_uri=${redirectUri}&scope=${scopes}&state=${state}`;
	//	}
	//
	//	function twitchLogout() {
	//		eraseCookie("accessToken");
	//		eraseCookie("refreshToken");
	//		window.location.href = "/";
	//	}
	//
	//	function isLoggedIn() {
	//		return getCookie("accessToken") !== null;
	//	}
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
					<th>Channel cooldown</th>
					<th>User cooldown</th>
					<th>No prefix</th>
					<th>Can disable</th>
				</tr>
			</thead>
			<tbody>
				{#each commands as cmd}
					<tr>
						<td>{cmd.Name}</td>
						<td>{cmd.Usage}</td>
						<td>{cmd.Description}</td>
						<td>{cmd.ChannelCooldown}</td>
						<td>{cmd.UserCooldown}</td>
						<td>
							<input
								type="checkbox"
								disabled
								class="checkbox"
								checked={cmd.NoPrefix}
							/>
						</td>
						<td>
							<input
								type="checkbox"
								disabled
								class="checkbox"
								checked={cmd.CanDisable}
							/>
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
</Container>
