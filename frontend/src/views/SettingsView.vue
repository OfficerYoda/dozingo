<template> 
  <section>
    <div class="grid container" style="margin-bottom: 1%;">
      <article class="card col-12 md-12">
        <div v-if="!auth.state.ready">Loading...</div>
        <div v-else-if="!auth.state.user">Not logged in.</div>
        <div v-else class="profile-header">
          <div class="avatar-wrapper" @click="fileInput?.click()">
            <img :src="auth.state.user.avatar_url ?? '/user.png'" class="profile-avatar" />
            <div class="avatar-overlay">
              <img src="/camera.png" alt="Change picture" class="camera-icon" />
            </div>
            <input ref="fileInput" type="file" accept="image/png,image/jpeg,image/webp"
              style="display:none" @change="uploadAvatar" />
          </div>
          <div class="username-row">
            <h1 v-if="!editingUsername">Welcome, {{ auth.state.user.username }}!</h1>
            <input v-else class="username-input" v-model="newUsername" @keyup.enter="saveUsername" @keyup.escape="editingUsername = false" autofocus />
            <div class="pencil-overlay" @click="startEditing">
              <img src="/pencil.png" alt="Change username" class="pencil-icon" />
            </div>
          </div>
        </div>
      </article>
    </div>

    <div class="grid container">
      <article class="card col-6 md-12">
        <div class="account-security-title">
          <ShieldUser :size="30" />
          <h2 class="mb-0">{{ $t('settings.security.title') }}</h2>
        </div>

        <span class="account-security-subtitle">{{ $t('settings.security.subtitle') }}</span>

        <div class="highlighedcard">
          <Smartphone :size="23" />
          <div class="account-security-info">
            <span>{{ $t('settings.security.twoFa') }}</span>
            <small>{{ $t('settings.security.disabled') }}</small>
          </div>
          <button class="btn btn-primary">{{ $t('settings.security.enable') }}</button>
        </div>

        <div class="highlighedcard">
          <Key :size="23" />
          <div class="account-security-info">
            <span>{{ $t('settings.security.lastPasswordChange') }}</span>
            <small>3 Months ago</small>
          </div>
          <button class="btn btn-primary">{{ $t('settings.security.change') }}</button>
        </div>

        <div class="display-darkmode">
          <Bell :size="23" />
          <div class="display-darkmode-info">
            <span>{{ $t('settings.security.loginNotification') }}</span>
            <small>{{ $t('settings.security.loginNotificationDesc') }}</small>
          </div>
          <label class="toggle">
            <input type="checkbox" id="btnToggle" name="btnToggle" />
            <span class="slider"></span>
          </label>
        </div>
      </article>

      <article class="card col-6 md-12">
        <div class="display-title">
          <Palette :size="30" />
          <h2 class="mb-0">{{ $t('settings.display.title') }}</h2>
        </div>

        <span class="account-security-subtitle">{{ $t('settings.display.subtitle') }}</span>

        <div class="display-darkmode">
          <Moon :size="23" />
          <div class="display-darkmode-info">
            <span>{{ $t('settings.display.darkMode') }}</span>
            <small>{{ $t('settings.display.darkModeDesc') }}</small>
          </div>
          <label class="toggle">
            <input type="checkbox" id="btnDarkToggle" name="btnDarkToggle" v-model="isChecked"
              @change="changeDarkMode" />
            <span class="slider"></span>
          </label>
        </div>

        <div class="display-darkmode">
          <Eye :size="23" />
          <div class="display-darkmode-info">
            <span>{{ $t('settings.display.colorCorrection') }}</span>
            <small>{{ $t('settings.display.colorCorrectionDesc') }}</small>
          </div>

          <select v-model="colorCorrection" class="btn btn-secondary">
            <option value="standart">{{ $t('settings.display.colorFilters.standard') }}</option>
            <option value="redgreen">{{ $t('settings.display.colorFilters.redGreen') }}</option>
            <option value="blueyellow">{{ $t('settings.display.colorFilters.blueYellow') }}</option>
            <option value="gray">{{ $t('settings.display.colorFilters.grayscale') }}</option>
          </select>
        </div>

        <div class="display-fontsize">
          <div class="display-fontsize-header">
            <ALargeSmall :size="23" />
            <small class="display-fontsize-title">{{ $t('settings.display.fontSize') }}</small>
          </div>
          <input class="sliderreal" type="range" min="1" max="5" step="1" v-model="fontSize">
          <div class="display-fontsize-texts">
            <small style="font-size: 12px;">{{ $t('settings.display.small') }}</small>
            <small style="font-size: 16px;">{{ $t('settings.display.standard') }}</small>
            <small style="font-size: 20px;">{{ $t('settings.display.large') }}</small>
          </div>
        </div>

      </article>

      <article class="card deactivateaccount col-12">
        <div class="deactivateaccount-title">
          <div class="deactivateaccount-heading">
            <UserX :size="23" />
            <h2 class="mb-0">{{ $t('settings.account.deactivate') }}</h2>
          </div>
          <small>{{ $t('settings.account.deactivateDesc') }}</small>
        </div>
        <button class="btn btn-danger deactivateaccount-btn">{{ $t('settings.account.deactivateBtn') }}</button>
      </article>
    </div>
  </section>
</template>

<script setup lang="ts">
import { useAuth } from '@/composables/useAuth'
import { ref, watch, useTemplateRef } from 'vue';
import { useI18n } from 'vue-i18n';

const auth = useAuth()
if (!auth.state.ready) {
  auth.fetchUser()
}

const editingUsername = ref(false)
const newUsername = ref('')

const fileInput = useTemplateRef<HTMLInputElement>('fileInput')

useI18n()

const isChecked = ref(localStorage.getItem('theme') === 'dark')
const colorCorrection = ref(localStorage.getItem('colorCorrection') ?? 'standart')
const fontSize = ref(Number(localStorage.getItem('fontSize') ?? 3))

const fontSizes = ['12px', '14px', '16px', '18px', '20px']

watch(fontSize, (val) => {
  document.documentElement.style.fontSize = fontSizes[val - 1] ?? '16px'
  localStorage.setItem('fontSize', String(val))
})



watch(colorCorrection, (newValue) => {
  localStorage.setItem('colorCorrection', newValue)
  if (newValue === 'standart') {
    document.documentElement.removeAttribute('color-correction')
  } else {
    document.documentElement.setAttribute('color-correction', newValue)
  }
}, { immediate: true })

function changeDarkMode() {
  if (isChecked.value) {
    document.documentElement.setAttribute('data-theme', 'dark')
    localStorage.setItem('theme', 'dark')
  } else {
    document.documentElement.removeAttribute('data-theme')
    localStorage.setItem('theme', 'light')
  }
}

async function uploadAvatar(event: Event) {
    const file = (event.target as HTMLInputElement).files?.[0]
    if (!file) return
    const form = new FormData()
    form.append('avatar', file)
    const res = await fetch('/api/users/me/avatar', {
        method: 'PUT',
        credentials: 'include',
        body: form
    })
    if (res.ok) {
        window.location.reload()
    }
}

function startEditing() {
  newUsername.value = auth.state.user?.username ?? ''
  editingUsername.value = true
}

async function saveUsername() {
  if (!newUsername.value.trim()) return
  const res = await fetch('/api/users/me', {
    method: 'PATCH',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username: newUsername.value })
  })
  if (res.ok) {
    editingUsername.value = false
    await auth.fetchUser()
  }
}

</script>

<style scoped>

.username-row {
  display: flex;
  align-items: center;
  gap: 0.1rem;
}

.card {
  background-color: var(--color-bg-card-tinted);
}

.card.deactivateaccount {
  background-color: var(--color-accent-red-soft);
}

.highlighedcard {
  background-color: var(--color-bg-surface);
  padding: 24px;
  margin-top: 24px;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  gap: 12px;
}

.highlighedcard .account-security-info {
  display: flex;
  flex-direction: column;
  flex: 1;
}

.display-title {
  display: flex;
  flex-direction: row;
  gap: 8px;
}

.profile-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 1rem;
  margin-top: 1rem;
}

.profile-header h1 {
  margin: 0;
}

.profile-avatar {
  width: 3rem;
  height: 3rem;
  border-radius: 50%;
  object-fit: cover;
  flex-shrink: 0;
}

.avatar-wrapper {
  position: relative;
  cursor: pointer;
  width: 3rem;
  height: 3rem;
}

.avatar-overlay {
  position: absolute;
  bottom: 0;
  right: 0;
  border-radius: 50%;
  background: rgba(0,0,0,0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0.2rem;
  opacity: 0;
  transition: opacity 0.2s;
  color: white;
}

.pencil-overlay {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0.2rem;
  cursor: pointer;
}

.avatar-wrapper:hover .avatar-overlay {
  opacity: 1;
}

.camera-icon {
  width: 1.2rem;
  height: 1.2rem;
  object-position: 3rem, 0.5rem;
}

.pencil-icon {
  width: 1rem;
  height: 1rem;
}

.display-title svg {
  color: var(--card-yellow);
}

.account-security-info span {
  color: var(--color-heading);
  font-size: 16px;
  font-weight: 550;
}

.account-security-info small {
  color: var(--color-text-subtle);
}

.display-darkmode span {
  color: var(--color-heading);
  font-size: 16px;
  font-weight: 550;
}

.display-darkmode small {
  color: var(--color-text-subtle);
}

.account-security-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.account-security-title svg {
  color: var(--card-red);
}

.display-darkmode {
  padding: 24px;
  margin-top: 24px;
  background-color: var(--color-bg-surface);
  border-radius: var(--radius-sm);
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: 12px;
}

.display-darkmode svg {
  color: var(--card-blue);
  flex-shrink: 0;
}

.display-darkmode-info {
  display: flex;
  flex-direction: column;
  flex: 1;
}

h2 {
  color: var(--color-heading);
}

.highlighedcard svg {
  color: var(--card-blue);
}

.account-security-subtitle {
  color: var(--color-text-subtle);
  font-weight: 500;
  padding-block-start: 20px;
}


.deactivateaccount {
  display: flex;
}

.deactivateaccount-heading {
  display: flex;
  align-items: center;
  gap: 8px;
}

.deactivateaccount-heading svg {
  color: var(--card-red);
}

.deactivateaccount-title {
  display: flex;
  flex-direction: column;
  flex: 1;
}

.deactivateaccount-btn {
  font-weight: 600;
}

/*--- Slider styling ---*/
input[type=range] {
  -webkit-appearance: none;
  appearance: none;
  width: 100%;
  background-color: #ffffff00;
}

input[type=range]::-webkit-slider-runnable-track {
  background-color: var(--color-interactive-track);
  border-radius: var(--radius-lg);
  width: 100%;
}

input[type=range]::-webkit-slider-thumb {
  -webkit-appearance: none;
  appearance: none;
  background: var(--color-text-subtle);
  margin-top: -5px;
  width: 20px;
  height: 20px;
  margin-bottom: -5px;
  border-radius: 50%;
}

.display-fontsize-texts {
  display: flex;
  justify-content: space-between;
}

.display-fontsize-texts small {
  color: var(--color-text-subtle);
  font-size: 12px;
  font-weight: 600;
}

.display-fontsize {
  margin-top: 24px;
}

.display-fontsize-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.display-fontsize-title {
  font-size: 16px;
  font-weight: 700;
}

.sliderreal {
  padding-top: 16px;
}




/* Das ist testing ground */
.toggle {
  position: relative;
  display: inline-block;
  width: 60px;
  height: 34px;
}

.toggle input {
  position: absolute;
  opacity: 0;
  width: 0;
  height: 0;
}

.toggle input:focus-visible+.slider {
  outline: 2px solid #4052B6;
  outline-offset: 2px;
}

.slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: var(--color-interactive-track);
  transition: 0.4s;
  border-radius: 34px;
}

.slider:before {
  position: absolute;
  content: "";
  height: 26px;
  width: 26px;
  left: 4px;
  bottom: 4px;
  background-color: #FFF;
  transition: 0.4s;
  border-radius: 50%;
}

input:checked+.slider {
  background-color: var(--color-text-subtle);
}

input:checked+.slider:before {
  transform: translateX(26px);
}
</style>
