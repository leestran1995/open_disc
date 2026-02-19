<script>
  import { onMount } from 'svelte';
  import { getUser, getUserRooms } from './lib/api.js';
  import { currentUser, rooms } from './lib/stores.js';
  import { connectSSE } from './lib/sse.js';
  import Login from './lib/Login.svelte';
  import Sidebar from './lib/Sidebar.svelte';
  import RoomHeader from './lib/RoomHeader.svelte';
  import MessageList from './lib/MessageList.svelte';
  import MessageInput from './lib/MessageInput.svelte';

  let ready = $state(false);

  onMount(async () => {
    const storedId = localStorage.getItem('user_id');
    if (storedId) {
      const user = await getUser(storedId);
      if (user) {
        currentUser.set(user);
        connectSSE(user.user_id);

        const userRooms = await getUserRooms(user.user_id);
        if (userRooms) {
          rooms.set(userRooms);
        }
      } else {
        localStorage.removeItem('user_id');
      }
    }
    ready = true;
  });
</script>

{#if !ready}
  <div class="loading">Loading...</div>
{:else if $currentUser}
  <div class="app-layout">
    <Sidebar />
    <main>
      <RoomHeader />
      <MessageList />
      <MessageInput />
    </main>
  </div>
{:else}
  <Login />
{/if}

<style>
  .loading {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100vh;
    color: var(--text-primary);
    opacity: 0.5;
  }

  .app-layout {
    display: grid;
    grid-template-columns: 240px 1fr;
    height: 100vh;
    overflow: hidden;
  }

  main {
    display: flex;
    flex-direction: column;
    height: 100vh;
    overflow: hidden;
  }
</style>
