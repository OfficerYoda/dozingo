<template>
  <div>
    <header>
      <h3>Settings</h3>
    </header>

    <main>
      <ul class="accountsettings">
        <li class="title">
          <ShieldUser :size="30"/>
          <span>Account Security</span>
        </li>

        <li class="subtitle">Authentication</li>

        <li class="highlighedcard">
          <Smartphone :size="23"/>
          <div class="info">
            <span>2FA Security</span>
            <small>Currently Disabled</small>
          </div>
          <button>Enable</button>
        </li>

        <li class="highlighedcard">
          <Key :size="23"/>
          <div class="info">
            <span>Last password change</span>
            <small>3 Months ago</small>
          </div>
          <button>Change</button>
        </li>
      </ul>

      <div class="display card">
        <span>Display</span>
      </div>

      <div class="deactivateaccount card">
        <span>Deactive Account</span>
        <small>Temporarily hide your profile and rankings from the playground.</small>
        <button class="btn btn-primary">Deactivate</button>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ShieldUser, Key, Smartphone } from 'lucide-vue-next'
</script>

<style scoped>

header {
  background-color: #aca8d718;
  padding: 16px 24px;
  width: 100%;
  display: flex;
  align-items: center;
}

h3 {
  margin: 0;
  color: var(--color-primary-800);
}

main {
  padding: 24px;
  display: flex;
  flex-direction: row;
  flex-wrap: wrap;
}

.card {
  flex: 1;
}

.accountsettings {
  background-color: #F3EEFF;
  border-radius: var(--radius-sm);
  padding: 24px;
}

.highlighedcard {
  background-color: #E9E5FF;
  width: 500px;
  padding: 24px;
  border-radius: var(--radius-sm);
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: 12px;
}

.highlighedcard .info {
  display: flex;
  flex-direction: column;
  flex: 1;
}

.info span {
  color: #2C2A51;
  font-size: 16px;
  font-weight: 550;
}

.info small {
  color: #75729E;
}

ul {
  list-style-type: none;
}

.title{
  display: flex;
  flex-direction: row;
}

.title svg {
  color: #B41340;
}

.title span {
  color: #2C2A51;
  font-size: 19px;
  font-weight: 600;
}

.highlighedcard svg {
  color: #4052B6;
}

.subtitle {
  color: #75729E;
  font-weight: 500;
  padding-block-start: 20px;
  padding-block-end: 13px;
}

</style>