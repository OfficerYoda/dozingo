import { createRouter, createWebHistory } from 'vue-router'
import { useAuth } from '@/composables/useAuth'

import HomeView from '@/views/HomeView.vue'
import LoginView from '@/views/LoginView.vue'
import SettingsView from '@/views/SettingsView.vue'
import ProfileView from '@/views/ProfileView.vue'
import ComponentsView from '@/views/ComponentsView.vue'
import CardEditorView from '@/views/CardEditorView.vue'
import CardsView from '@/views/CardsView.vue'
import RegisterView from '@/views/RegisterView.vue'
import ForgotPasswordView from '@/views/ForgotPasswordView.vue'

declare module 'vue-router' {
  interface RouteMeta {
    requiresAuth?: boolean
    guestOnly?: boolean
  }
}

const routes = [
  { path: '/', name: 'home', component: HomeView, meta: { requiresAuth: false } },
  { path: '/login', name: 'login', component: LoginView, meta: { guestOnly: true } },
  { path: '/register', name: 'register', component: RegisterView, meta: { guestOnly: true } },
  { path: '/forgotpw', name: 'forgotpw', component: ForgotPasswordView, meta: { guestOnly: true } },
  { path: '/settings', name: 'settings', component: SettingsView, meta: { requiresAuth: false } },
  { path: '/profile', name: 'profile', component: ProfileView, meta: { requiresAuth: true } },
  { path: '/components', name: 'components', component: ComponentsView, meta: { requiresAuth: false } },
  { path: '/cardeditor', name: 'cardeditor', component: CardEditorView, meta: { requiresAuth: true } },
  { path: '/cards', name: 'cards', component: CardsView, meta: { requiresAuth: false } },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

router.beforeEach(async (to) => {
  const { state, fetchUser } = useAuth()
  if (!state.ready) {
    await fetchUser()
  }
  if (to.meta.requiresAuth && !state.user) {
    return { name: 'login' }
  }
  if (to.meta.guestOnly && state.user) {
    return { name: 'home' }
  }
})

export default router
