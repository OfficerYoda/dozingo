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
            <li><RouterLink to="/cardeditor" class="sidebar-buttons upper-buttons"><SquarePen :size="20"/>{{ $t('nav.cardEditor') }}</RouterLink></li>
            <li><RouterLink to="/boards" class="sidebar-buttons upper-buttons"><SquarePen :size="20"/>Boards</RouterLink></li>
            <li><RouterLink to="/components" class="sidebar-buttons upper-buttons"><Computer :size="20"/>Components(only for dev)</RouterLink></li>
          </ul>
        </nav>
      </div>

      <div class="sidebar-footer">
        <hr class="m-0">
        <nav>
          <ul>
            <li><RouterLink to="/settings" class="sidebar-buttons"><Settings :size="20"/>{{ $t('nav.settings') }}</RouterLink></li>
            <li v-if="auth.state.user">
              <button class="sidebar-buttons" @click="handleLogout"><LogOut :size="20" class="sidebar-footer-signout"/><span class="sidebar-footer-signout">{{ $t('nav.signOut') }}</span></button>
            </li>
            <li v-else>
              <RouterLink to="/login" class="sidebar-buttons"><LogIn :size="20" />Log In</RouterLink>
            </li>
            <hr class="m-0">
            <li><RouterLink class="sidebar-buttons" to="/profile"><UserCircle :size="20"/>{{ auth.state.user?.username ?? $t('nav.profile') }}</RouterLink></li>
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
          <h1 class="mb-0">{{ pageTitle }}</h1>
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

        <footer class="text-muted">
          {{ $t('footer.copyright') }}
        </footer>

      <div class="disable-layer" id="disable-layer" @click="closeMenu"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Home, Settings, LogOut, LogIn, UserCircle, Menu, Computer, SquarePen, Languages } from 'lucide-vue-next'
import { useAuth } from '@/composables/useAuth'
import { usePageTitle } from '@/composables/usePageTitle'

const { locale } = useI18n()
const router = useRouter()
const auth = useAuth()
const { pageTitle } = usePageTitle()

const fontSizes = ['12px', '14px', '16px', '18px', '20px']
const savedFontSize = localStorage.getItem('fontSize')
if (savedFontSize) {
  document.documentElement.style.fontSize = fontSizes[Number(savedFontSize) - 1] ?? '16px'
}

watch(locale, (newLocale) => {
  localStorage.setItem('locale', newLocale)
})

async function handleLogout() {
  await auth.logout()
  router.push('/login')
}
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