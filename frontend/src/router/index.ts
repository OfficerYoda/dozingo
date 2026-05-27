import { createRouter, createWebHistory } from 'vue-router'

import HomeView from '@/views/HomeView.vue'
import LoginView from '@/views/LoginView.vue'
import SettingsView from '@/views/SettingsView.vue'
import ProfileView from '@/views/ProfileView.vue'
import ComponentsView from '@/views/ComponentsView.vue'
import CardEditorView from '@/views/CardEditorView.vue'
import CardsView from '@/views/CardsView.vue'
import RegisterView from '@/views/RegisterView.vue'


const routes = [
  { path: '/', name: 'home', component: HomeView },
  { path: '/login', name: 'login', component: LoginView },
  { path: '/register', name: 'register', component: RegisterView },
  { path: '/settings', name: 'settings', component: SettingsView },
  { path: '/profile', name: 'profile', component: ProfileView },
  { path: '/components', name: 'components', component: ComponentsView },
  { path: '/cardeditor', name: 'cardeditor', component: CardEditorView },
  { path: '/cards', name: 'cards', component: CardsView },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

export default router