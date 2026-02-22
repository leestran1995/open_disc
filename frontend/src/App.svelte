<script>
  import { onMount } from 'svelte';
  import { currentUser, authToken, rooms, activeRoomId } from './lib/stores.js';
  import { connectSSE } from './lib/sse.js';
  import { decodeJWT } from './lib/jwt.js';
  import { getRooms } from './lib/api.js';
  import Login from './lib/Login.svelte';
  import Sidebar from './lib/Sidebar.svelte';
  import RoomHeader from './lib/RoomHeader.svelte';
  import MessageList from './lib/MessageList.svelte';
  import MessageInput from './lib/MessageInput.svelte';

  let ready = $state(false);

  onMount(() => {
    const token = localStorage.getItem('token');
    if (token) {
      const claims = decodeJWT(token);
      if (claims && claims.exp > Date.now() / 1000) {
        const username = claims.username;
        authToken.set(token);
        currentUser.set({ username });

        const storedRooms = localStorage.getItem('rooms');
        if (storedRooms) {
          try {
            rooms.set(JSON.parse(storedRooms));
          } catch { /* ignore bad data */ }
        }

        const storedRoomId = localStorage.getItem('activeRoomId');
        if (storedRoomId) {
          activeRoomId.set(storedRoomId);
        }

        connectSSE(token, username);

        getRooms().then((result) => {
          if (Array.isArray(result)) {
            rooms.set(result);
            localStorage.setItem('rooms', JSON.stringify(result));
          }
        });
      } else {
        localStorage.removeItem('token');
      }
    }
    ready = true;

    return activeRoomId.subscribe((id) => {
      if (id) {
        localStorage.setItem('activeRoomId', id);
      } else {
        localStorage.removeItem('activeRoomId');
      }
    });
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
    min-width: 0;
  }
</style>
