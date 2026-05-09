<template>
  <div class="app-layout">
      <header>
    <h1><!-- Page Title --></h1>
    <!-- Sidebar open btn later on -->
  </header>

  <aside>
    <div class="sidebar-header">
      <span>DOZINGO</span>
      <!-- Sidebar close btn later on -->
    </div>

    <div class="sidebar-content">
      <nav>
        <ul>
          <li><RouterLink class="btn" to="/">Home</RouterLink></li>
          <li><RouterLink class="btn btn-primary" to="/about">Kategorien</RouterLink></li>
        </ul>
      </nav>
    </div>

    <div class="sidebar-footer">
      <nav>
        <ul>
          <li><RouterLink to="/settings">Settings</RouterLink></li>
          <li class="sidebar-footer-signout">Sign Out</li>
          <li><RouterLink to="/profile">Profile</RouterLink></li>
        </ul>
      </nav>
    </div>

  </aside>

  <main>
    <RouterView />
  </main>

  <footer>
    &#169 DOZINGO - Alle Rechte vorbehalten
  </footer>
  </div>
</template>

<script setup lang="ts">

</script>

<style scoped>

.app-layout {
  display: grid;
  height: 100vh;
  grid-template-columns: 250px 1fr;
  grid-template-rows: auto 1fr auto;
}

nav li {
  list-style-type:none;
  color: var(--color-primary-600);
  font-size: 0.875rem;
  font-weight: 700;
}

header {
  grid-column: 2;
  grid-row: 1;
}

aside {
  background-color: var(--color-bg-sidebar);
  grid-column: 1;
  grid-row: 1 / -1;
}

main {
  grid-column: 2;
  grid-row: 1;
}

footer {
  grid-column: 2;
  grid-row: 3;
}

.sidebar-footer-signout {
  color: var(--color-accent-red)
}

</style>