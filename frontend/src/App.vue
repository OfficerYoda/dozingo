<template>
  <div class="app-layout">
    <header>
      <h1><!-- Page Title --></h1>
    </header>

    <aside>
      <div class="sidebar-header">
        <span class="brand-title">DOZINGO</span>
        <small class="brand-subtitle">Schoolary Playground</small>
      </div>

      <div class="sidebar-content">
        <nav>
          <ul>
            <li><RouterLink to="/" class="sidebar-buttons upper-buttons"><Home :size="20" />Home</RouterLink></li>
            <li><RouterLink to="/about" class="sidebar-buttons upper-buttons"><LayoutGrid :size="20"/>About</RouterLink></li>
          </ul>
        </nav>
      </div>

      <div class="sidebar-footer">
        <hr class="divider">
        <nav>
          <ul>
            <li><RouterLink to="/settings" class="sidebar-buttons"><Settings :size="20"/>Settings</RouterLink></li>
            <li><button class="sidebar-buttons"><LogOut :size="20" class="sidebar-footer-signout"/><span class="sidebar-footer-signout">Sign Out</span></button></li>
            <hr class="divider">
            <li><RouterLink class="sidebar-buttons" to="/profile"><UserCircle :size="20"/>Profile</RouterLink></li>
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

.brand-title {
  display: block;
  font-size: 2.5rem;
  font-weight: 700;
  color: var(--color-primary-800);
}

.brand-subtitle {
  display: block;
  font-size: 0.875rem;
}

header {
  grid-column: 2;
  grid-row: 1;
}

aside {
  padding: 20px;
  background-color: var(--color-bg-sidebar);
  grid-column: 1;
  grid-row: 1 / -1;
  display: grid;
  grid-template-rows: auto 1fr auto;
}

.sidebar-header {
  grid-row: 1;
  margin-bottom: 30px;
}

.sidebar-content {
  grid-row: 2;
}

.sidebar-footer {
  grid-row: 3;
}

main {
  grid-column: 2;
  grid-row: 2;
}

footer {
  grid-column: 2;
  grid-row: 3;
}

.sidebar-buttons {
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: 8px;
  padding: 10px;
  font-weight: 400;
}

.sidebar-footer-signout {
  color: var(--color-accent-red);
}

.router-link-exact-active.upper-buttons {
  background-color: var(--color-primary-200);
  border-radius: var(--radius-sm);
}

.divider {
  border: none;
  border-top: 1px solid var(--color-primary-100);
}

button {
  background: none;
  border: none;
  padding: 0;
  cursor: pointer;
  font: inherit;
}

</style>