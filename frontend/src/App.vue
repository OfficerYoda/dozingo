<template>
  <div class="app-layout">
      <header>
    <h1><!-- Page Title --></h1>
  </header>

  <aside>
    <div class="sidebar-header">
      <span>DOZINGO</span>
    </div>

    <div class="sidebar-content">
      <nav>
        <ul>
          <li><RouterLink to="/" class="sidebar-buttons"><Home :size="16" /> Home</RouterLink></li>
          <li><RouterLink to="/about" class="sidebar-buttons"><LayoutGrid :size="16"/>Kategorien</RouterLink></li>
        </ul>
      </nav>
    </div>

    <div class="sidebar-footer">
      <nav>
        <ul>
          <li><RouterLink class="sidebar-buttons" to="/settings"><Settings :size="16"/>Settings</RouterLink></li>
          <li class="sidebar-buttons sidebar-footer-signout"><LogOut :size="16" /> Sign Out</li>
          <li><RouterLink class="sidebar-buttons" to="/profile"><UserCircle :size="16" /> Profile</RouterLink></li>
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
import { Home, LayoutGrid, Settings, LogOut, UserCircle} from 'lucide-vue-next'
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
  display: grid;
  grid-template-rows: auto 1fr auto;
}

.sidebar-header {
  grid-row: 1;
  padding-inline: 10px;
}

.sidebar-content {
  grid-row: 2;
  padding-inline: 10px;
}

.sidebar-footer {
  grid-row: 3;
  padding-inline: 10px;
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
  color: var(--color-accent-red);
}

.sidebar-buttons {
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: 8px;
}

</style>