<template>
  <section>
  <div class="grid container">
    <article class="accountsettings col-6 md-12">
      <div class="title">
        <ShieldUser :size="30"/>
        <h2 class="mb-0">Account Security</h2>
      </div>

      <span class="subtitle">Authentication</span>

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

    <article class="display card col-6 md-12">
      <span>Display</span>
    </article>

    <article class="deactivateaccount card col-12">
      <span>Deactive Account</span>
      <small>Temporarily hide your profile and rankings from the playground.</small>
      <button class="btn btn-primary">Deactivate</button>
    </article>
  </div>
  </section>
</template>

<script setup lang="ts">
import { ShieldUser, Key, Smartphone } from 'lucide-vue-next'
</script>

<style scoped>

.accountsettings {
  background-color: #F3EEFF;
  border-radius: var(--radius-sm);
  padding: 24px;
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
}

</style>