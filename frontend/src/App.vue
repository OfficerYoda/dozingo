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
            <li><RouterLink to="/boards" class="sidebar-buttons upper-buttons"><LayoutGrid :size="20"/>{{ $t('nav.boards') }}</RouterLink></li>
            <li><RouterLink to="/components" class="sidebar-buttons upper-buttons"><Computer :size="20"/>Components(only for dev)</RouterLink></li>
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
          <div class="profile-menu" ref="profileMenuRef">
            <button class="profile-trigger" @click="profileOpen = !profileOpen">
              <img
                v-if="auth.state.user?.avatar_url"
                :src="auth.state.user.avatar_url"
                class="profile-avatar"
              />
              <UserCircle v-else :size="28" />
            </button>
            <div v-if="profileOpen" class="profile-dropdown">
              <template v-if="auth.state.user">
                <RouterLink to="/profile" class="dropdown-item" @click="profileOpen = false">
                  <UserCircle :size="16" /> {{ auth.state.user.username }}
                </RouterLink>
                <RouterLink to="/settings" class="dropdown-item" @click="profileOpen = false">
                  <Settings :size="16" /> {{ $t('nav.settings') }}
                </RouterLink>
                <hr class="dropdown-divider">
                <button class="dropdown-item dropdown-item-danger" @click="handleLogout">
                  <LogOut :size="16" /> {{ $t('nav.signOut') }}
                </button>
              </template>
              <template v-else>
                <button class="dropdown-item" @click="openLoginModal(); profileOpen = false">
                  <LogIn :size="16" /> {{ $t('nav.signIn') }}
                </button>
                <button class="dropdown-item" @click="openRegisterModal(); profileOpen = false">
                  <UserPlus :size="16" /> {{ $t('nav.register') }}
                </button>
                <hr class="dropdown-divider">
                <RouterLink to="/settings" class="dropdown-item" @click="profileOpen = false">
                  <Settings :size="16" /> {{ $t('nav.settings') }}
                </RouterLink>
              </template>
            </div>
          </div>
        </div>
      </header>

      <main>
        <RouterView />
      </main>

        <footer class="text-muted">
          {{ $t('footer.copyright') }}
        </footer>

      <LoginModal />
      <RegisterModal />
      <div class="disable-layer" id="disable-layer" @click="closeMenu"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Home, Settings, LogOut, LogIn, UserCircle, UserPlus, Menu, Computer, SquarePen, LayoutGrid } from 'lucide-vue-next'
import { useAuth } from '@/composables/useAuth'
import { useLoginModal } from '@/composables/useLoginModal'
import { useRegisterModal } from '@/composables/useRegisterModal'
import LoginModal from '@/components/LoginView.vue'
import RegisterModal from '@/components/RegisterView.vue'
import { usePageTitle } from '@/composables/usePageTitle'

const { openLoginModal } = useLoginModal()
const { openRegisterModal } = useRegisterModal()
const { locale } = useI18n()
const router = useRouter()
const auth = useAuth()
const { pageTitle } = usePageTitle()
const profileOpen = ref(false)
const profileMenuRef = ref<HTMLElement | null>(null)

function handleClickOutside(e: MouseEvent) {
  if (profileMenuRef.value && !profileMenuRef.value.contains(e.target as Node)) {
    profileOpen.value = false
  }
}

onMounted(() => document.addEventListener('click', handleClickOutside))
onUnmounted(() => document.removeEventListener('click', handleClickOutside))

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
  router.push('/')
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

.profile-menu {
  position: relative;
  cursor: pointer;
}

.profile-trigger {
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-primary-600);
}

.profile-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  object-fit: cover;
  display: block;
}

.profile-dropdown {
  position: absolute;
  top: calc(100% + 8px);
  right: 0;
  background: var(--color-bg-card-tinted);
  border: 1px solid var(--color-border, #e2e8f0);
  border-radius: var(--radius-sm);
  min-width: 180px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
  z-index: 100;
  overflow: hidden;
}

.dropdown-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  font-size: 0.875rem;
  width: 100%;
  text-align: left;
  background: none;
  border: none;
  cursor: pointer;
  color: var(--color-primary-600);
  text-decoration: none;
  font: inherit;
}

.dropdown-item:hover {
  background-color: var(--color-primary-200);
}

.dropdown-item-danger {
  color: var(--color-accent-red);
}

.dropdown-divider {
  margin: 0;
  border-color: var(--color-border, #e2e8f0);
}
</style>