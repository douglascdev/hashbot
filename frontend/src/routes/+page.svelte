<script>
  import { onMount } from 'svelte';
  let commands = [];

  onMount(async () => {
    
    const res = await fetch(`${import.meta.env.BASE_URL}commands.json`);
    commands = await res.json();
  });

  function twitchRedirect () {
		let state = Math.random().toString(36).slice(2);
		let redirectUri = 'http://localhost:8080/login';
		let clientId = 'y2dazd63f0f6fr8a8hbx2l0ibyqfyd';
		let scopes = 'channel%3Abot'
		document.cookie = `loginState=${state}`;
		window.location.href = `https://id.twitch.tv/oauth2/authorize?response_type=code&client_id=${clientId}&redirect_uri=${redirectUri}&scope=${scopes}&state=${state}`;
  }
</script>

<div class="navbar bg-base-100 shadow-sm">
  <div class="flex-1">
    <a class="btn btn-ghost text-xl">hashbot</a>
  </div>
  <button class="btn" on:click={twitchRedirect()}>
    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-twitch" viewBox="0 0 16 16">
      <path d="M3.857 0 1 2.857v10.286h3.429V16l2.857-2.857H9.57L14.714 8V0zm9.714 7.429-2.285 2.285H9l-2 2v-2H4.429V1.143h9.142z"/>
      <path d="M11.857 3.143h-1.143V6.57h1.143zm-3.143 0H7.571V6.57h1.143z"/>
    </svg>
    Login with Twitch
  </button>
</div>

<div class="overflow-x-auto rounded-box border border-base-content/5 bg-base-100">
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
          <input type="checkbox" disabled class="checkbox" checked={cmd.NoPrefix} />
        </td>
        <td>
          <input type="checkbox" disabled class="checkbox" checked={cmd.CanDisable} />
        </td>
      </tr>
    {/each}
  </tbody>
</table>
</div>

