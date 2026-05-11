<template>
  <div class="app-layout">
    <aside>
      <div class="sidebar-header">
        <span class="brand-title">DOZINGO</span>
        <span class="brand-subtitle">Schoolary Playground</span>
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

    <div class="content-area">
      <header>
        <h1><!-- Page Title --></h1>
      </header>

      <main>
        <RouterView />
      </main>

      <footer>
        &#169 DOZINGO - Alle Rechte vorbehalten
      </footer>
    </div>
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

aside {
  padding: 20px;
  background-color: var(--color-bg-sidebar);
  display: grid;
  grid-template-rows: auto 1fr auto;
  overflow: hidden;
}

.content-area {
  display: flex;
  flex-direction: column;
  overflow-y: auto;
}

main {
  flex: 1;
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