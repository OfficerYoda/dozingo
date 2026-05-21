<template>
  <section>
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
          <div class="display-darkmode-info">
            <span>{{ $t('settings.display.darkMode') }}</span>
            <small>{{ $t('settings.display.darkModeDesc') }}</small>
          </div>
          <label class="toggle">
            <input type="checkbox" id="btnDarkToggle" name="btnDarkToggle" v-model="isChecked" @change="changeDarkMode"/>
            <span class="slider"></span>
          </label>
        </div>

        <div class="display-darkmode">
          <div class="display-darkmode-info">
            <span>{{ $t('settings.display.colorCorrection') }}</span>
            <small>{{ $t('settings.display.colorCorrectionDesc') }}</small>
          </div>
          <div class="dropdown">
            <button class="btn btn-secondary dropdown-button">{{ $t('settings.display.colorTheme') }}</button>
            <ul class="dropdown-menu">
              <li>{{ $t('settings.display.colorFilters.redGreen') }}</li>
              <li>{{ $t('settings.display.colorFilters.blueYellow') }}</li>
              <li>{{ $t('settings.display.colorFilters.grayscale') }}</li>
            </ul>
          </div>
        </div>

        <div class="display-fontsize">
          <small class="display-fontsize-title">{{ $t('settings.display.fontSize') }}</small>
          <input class="sliderreal" type="range" min="1" max="5" value="3" step="1">
          <div class="display-fontsize-texts">
            <small>{{ $t('settings.display.small') }}</small>
            <small>{{ $t('settings.display.standard') }}</small>
            <small>{{ $t('settings.display.large') }}</small>
          </div>
        </div>

      </article>

      <article class="card deactivateaccount col-12">
        <div class="deactivateaccount-title">
          <h2 class="mb-0">{{ $t('settings.account.deactivate') }}</h2>
          <small>{{ $t('settings.account.deactivateDesc') }}</small>
        </div>
        <button class="btn btn-danger deactivateaccount-btn">{{ $t('settings.account.deactivateBtn') }}</button>
      </article>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { ShieldUser, Key, Smartphone, Palette } from 'lucide-vue-next';

useI18n()

const isChecked = ref(document.documentElement.getAttribute('data-theme') === 'dark')

function changeDarkMode() {
  if (isChecked.value) {
    document.documentElement.setAttribute('data-theme', 'dark');
  } else {
    document.documentElement.removeAttribute('data-theme');
  }
}

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

.display-title svg {
  color: #874E00;
}

.account-security-info span {
  color: #2C2A51;
  font-size: 16px;
  font-weight: 550;
}

.account-security-info small {
  color: #75729E;
}

.display-darkmode span {
  color: #2C2A51;
  font-size: 16px;
  font-weight: 550;
}

.display-darkmode small {
  color: #75729E;
}

.account-security-title {
  display: flex;
  gap: 8px;
}

.account-security-title svg {
  color: #B41340;
}

.display-darkmode {
  padding: 24px;
  margin-top: 24px;
  background-color: #FFFFFF;
  border-radius: var(--radius-sm);
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
  background-color: #DDD9FF;
  border-radius: var(--radius-lg);
  width: 100%;
}

input[type=range]::-webkit-slider-thumb {
  -webkit-appearance: none;
  appearance: none;
  background: #75729E;
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
  color: #75729E;
  font-size: 12px;
  font-weight: 600;
}

.display-fontsize {
  margin-top: 24px;
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
  background-color: #DDD9FF;
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
  background-color: #75729E;
}

input:checked+.slider:before {
  transform: translateX(26px);
}
</style>
