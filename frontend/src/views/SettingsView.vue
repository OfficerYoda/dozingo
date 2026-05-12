<template>
  <section>
  <div class="grid container">
    <article class="card col-6 md-12">
      <div class="account-security-title">
        <ShieldUser :size="30"/>
        <h2 class="mb-0">Account Security</h2>
      </div>

      <span class="account-security-subtitle">Authentication</span>

      <div class="highlighedcard">
        <Smartphone :size="23"/>
        <div class="info">
          <span>2FA Security</span>
          <small>Currently Disabled</small>
        </div>
        <button>Enable</button>
      </div>

      <div class="highlighedcard">
        <Key :size="23"/>
        <div class="info">
          <span>Last password change</span>
          <small>3 Months ago</small>
        </div>
        <button>Change</button>
      </div>
    </article>

    <article class="card col-6 md-12">
      <div class="display-title">
        <Palette :size="30"/>
        <h2 class="mb-0">Display</h2>
      </div>

      <div class="display-darkmode">
        <div class="display-darkmode-info">
          <span>Dark-Mode</span>
          <small>Adjust for night study</small>
        </div>
        <input type="checkbox" id="my-toggle">
      </div>

    </article>

    <article class="card deactivateaccount col-12">
      <div class="deactivateaccount-title">
        <h2 class="mb-0">Deactive Account</h2>
        <small>Temporarily hide your profile and rankings from the playground.</small>
      </div>
      <button class="btn btn-danger">Deactivate</button>
    </article>
  </div>
  </section>
</template>

<script setup lang="ts">
import { ShieldUser, Key, Smartphone, Palette } from 'lucide-vue-next'
</script>

<style scoped>

.card {
  background-color: #F3EEFF
}

.card.deactivateaccount {
  background-color: var(--color-accent-red-soft);
}

.highlighedcard {
  background-color: #E9E5FF;
  padding: 24px;
  margin-top: 24px;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  gap: 12px;
}

.highlighedcard .info {
  display: flex;
  flex-direction: column;
  flex: 1;
}

.display-title {
  display: flex;
  flex-direction: row;
  gap: 8px;
}

.display-title svg {
  color: #874E00;
}

.info span {
  color: #2C2A51;
  font-size: 16px;
  font-weight: 550;
}

.info small {
  color: #75729E;
}

.account-security-title{
  display: flex;
    gap: 8px;
}

.account-security-title svg {
  color: #B41340;
}

.display-darkmode {
  display: flex;
  flex-direction: row;
  align-items: center;
}

.display-darkmode-info {
  display: flex;
  flex-direction: column;
  flex: 1;
}

h2 {
  color: #2C2A51;
}

.highlighedcard svg {
  color: #4052B6;
}

.account-security-subtitle {
  color: #75729E;
  font-weight: 500;
  padding-block-start: 20px;
}


.deactivateaccount {
  display: flex;
}

.deactivateaccount-title {
  display: flex;
  flex-direction: column;
  flex: 1;
}

</style>