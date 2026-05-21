<script lang="ts">
function openMenu(){
  let aside = document.getElementById("aside");
  aside?.classList.add('aside-open');
  let disable_layer = document.getElementById("disable-layer");
  disable_layer?.classList.add('disable-layer-active');
}
function closeMenu(){
  let aside = document.getElementById("aside");
  aside?.classList.remove('aside-open');
  let disable_layer = document.getElementById("disable-layer");
  disable_layer?.classList.remove('disable-layer-active');
}
</script>

<template>
  <div class="app-layout">
    <aside id="aside">
      <div class="sidebar-header">
        <span class="brand-title">DOZINGO</span>
        <small class="brand-subtitle">Schoolary Playground</small>
      </div>

      <div class="sidebar-content my-3">
        <nav>
          <ul>
            <li><RouterLink to="/" class="sidebar-buttons upper-buttons"><Home :size="20" />{{ $t('nav.home') }}</RouterLink></li>
            <li><RouterLink to="/cardeditor" class="sidebar-buttons upper-buttons"><SquarePen :size="20"/>Card Editor</RouterLink></li>
            <li><RouterLink to="/components" class="sidebar-buttons upper-buttons"><Computer :size="20"/>Components(only for dev)</RouterLink></li>
          </ul>
        </nav>
      </div>

      <div class="sidebar-footer">
        <hr class="m-0">
        <nav>
          <ul>
            <li><RouterLink to="/settings" class="sidebar-buttons"><Settings :size="20"/>{{ $t('nav.settings') }}</RouterLink></li>
            <li><button class="sidebar-buttons"><LogOut :size="20" class="sidebar-footer-signout"/><span class="sidebar-footer-signout">{{ $t('nav.signOut') }}</span></button></li>
            <hr class="m-0">
            <li><RouterLink class="sidebar-buttons" to="/profile"><UserCircle :size="20"/>{{ $t('nav.profile') }}</RouterLink></li>
          </ul>
        </nav>
      </div>
    </aside>

    <div class="content-area">
      <header>
        <button @click="openMenu">
          <Menu :size="20"/>
        </button> 
        <div class="header-spacing">
          <h1 class="mb-0">{{ $t('nav.settings') }}</h1>
          <div class="dropdown">
            <button class="btn"><Languages :size="20" /></button>
            <ul class="dropdown-menu">
              <li @click="locale = 'en'">English</li>
              <li @click="locale = 'de'">Deutsch</li>
            </ul>
          </div>
        </div>
      </header>

      <main>
        <RouterView />
      </main>

      <footer>
        &#169 DOZINGO - Alle Rechte vorbehalten
      </footer>

      <div class="disable-layer" id="disable-layer" @click="closeMenu"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { Home, Settings, LogOut, UserCircle, Menu, Computer, SquarePen, Languages } from 'lucide-vue-next'

const { locale } = useI18n()
</script>

<style scoped>

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

button {
  background: none;
  border: none;
  padding: 0;
  cursor: pointer;
  font: inherit;
}

header {
  background-color: #aca8d718;
  padding: 16px 24px;
  width: 100%;
  display: flex;
  align-items: center;
}

.header-spacing {
  display: flex;
  flex: 1;
  align-items: center;
  justify-content: space-between;
}
</style>