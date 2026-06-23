<template>
  <section>
    <div class="container">

      <div class="privacy-hero">
        <div class="hero-badge">{{ t('privacy.heroBadge') }}</div>
        <h1>{{ t('privacy.title') }}</h1>
        <p class="hero-sub">
          {{ t('privacy.heroSub') }}<br>
          <span class="hero-sub-muted">{{ t('privacy.heroSubMuted') }}</span>
        </p>
      </div>

      <ul class="privacy-checklist">

        <li class="privacy-item">
          <span class="item-icon">🏢</span>
          <div class="item-body">
            <h2>{{ t('privacy.items.responsible.title') }}</h2>
            <p>{{ t('privacy.items.responsible.content') }}<br>
              <a href="mailto:kontakt@dozingo.de">kontakt@dozingo.de</a></p>
          </div>
          <span class="item-badge">{{ t('privacy.items.responsible.badge') }}</span>
        </li>

        <li class="privacy-item">
          <span class="item-icon">📋</span>
          <div class="item-body">
            <h2>{{ t('privacy.items.data.title') }}</h2>
            <p>{{ t('privacy.items.data.content') }}</p>
            <p class="item-note">{{ t('privacy.items.data.note') }}</p>
          </div>
          <span class="item-badge">{{ t('privacy.items.data.badge') }}</span>
        </li>

        <li class="privacy-item">
          <span class="item-icon">🎯</span>
          <div class="item-body">
            <h2>{{ t('privacy.items.purpose.title') }}</h2>
            <p>{{ t('privacy.items.purpose.content') }}</p>
          </div>
          <span class="item-badge item-badge-green">{{ t('privacy.items.purpose.badge') }}</span>
        </li>

        <li class="privacy-item">
          <span class="item-icon">🍪</span>
          <div class="item-body">
            <h2>{{ t('privacy.items.cookies.title') }}</h2>
            <p>{{ t('privacy.items.cookies.content') }}</p>
            <p class="item-note">{{ t('privacy.items.cookies.note') }}</p>
          </div>
          <span class="item-badge item-badge-green">{{ t('privacy.items.cookies.badge') }}</span>
        </li>

        <li class="privacy-item">
          <span class="item-icon">🤝</span>
          <div class="item-body">
            <h2>{{ t('privacy.items.sharing.title') }}</h2>
            <p>{{ t('privacy.items.sharing.content') }}</p>
          </div>
          <span class="item-badge item-badge-green">{{ t('privacy.items.sharing.badge') }}</span>
        </li>

        <li class="privacy-item">
          <span class="item-icon">⚖️</span>
          <div class="item-body">
            <h2>{{ t('privacy.items.rights.title') }}</h2>
            <p>{{ t('privacy.items.rights.content') }}
              <a href="mailto:kontakt@dozingo.de">kontakt@dozingo.de</a></p>
          </div>
          <span class="item-badge">{{ t('privacy.items.rights.badge') }}</span>
        </li>

        <li class="privacy-item">
          <span class="item-icon">🗑️</span>
          <div class="item-body">
            <h2>{{ t('privacy.items.retention.title') }}</h2>
            <p>{{ t('privacy.items.retention.content') }}</p>
            <p class="item-note">{{ t('privacy.items.retention.note') }}</p>
          </div>
          <span class="item-badge">{{ t('privacy.items.retention.badge') }}</span>
        </li>

      </ul>

      <div class="privacy-outro">
        <span class="outro-stamp">{{ t('privacy.stamp') }}</span>
        <p>{{ t('privacy.outro') }}</p>
      </div>

    </div>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { usePageTitle } from '@/composables/usePageTitle'

const { t } = useI18n()
usePageTitle(t('privacy.title'))
</script>

<style scoped>
.privacy-hero {
  padding: 40px 0 32px;
  max-width: 620px;
}

.hero-badge {
  display: inline-block;
  background-color: var(--color-accent-green);
  color: #fff;
  font-size: 0.7rem;
  font-weight: 800;
  letter-spacing: 0.1em;
  padding: 3px 10px;
  border-radius: var(--radius-sm);
  margin-bottom: 12px;
}

.privacy-hero h1 {
  font-size: 2.4rem;
  font-weight: 900;
  color: var(--color-heading);
  margin: 0 0 10px;
}

.hero-sub {
  color: var(--color-text-secondary);
  line-height: 1.6;
  margin: 0;
}

.hero-sub-muted {
  font-size: 0.85rem;
  color: var(--color-text-muted);
}

.privacy-checklist {
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 10px;
  max-width: 720px;
  margin-bottom: 32px;
  padding: 0;
}

.privacy-item {
  display: flex;
  align-items: flex-start;
  gap: 16px;
  background-color: var(--color-bg-surface);
  border: 2px solid var(--color-input-bg);
  border-radius: var(--radius-lg);
  padding: 18px 20px;
  transition: border-color 0.2s, transform 0.15s;
}

.privacy-item:hover {
  border-color: var(--color-subheading);
  transform: translateX(4px);
}

.item-icon {
  font-size: 1.4rem;
  flex-shrink: 0;
  margin-top: 2px;
}

.item-body {
  flex: 1;
}

.item-body h2 {
  font-size: 0.85rem;
  font-weight: 700;
  color: var(--color-subheading);
  text-transform: uppercase;
  letter-spacing: 0.07em;
  margin: 0 0 6px;
}

.item-body p {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
  line-height: 1.6;
  margin: 0;
}

.item-body p + p {
  margin-top: 4px;
}

.item-note {
  color: var(--color-text-muted) !important;
  font-style: italic;
  font-size: 0.8rem !important;
}

.item-body a {
  color: var(--card-blue);
}

.item-badge {
  flex-shrink: 0;
  align-self: flex-start;
  background-color: var(--color-input-bg);
  color: var(--color-subheading);
  font-size: 0.6rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  padding: 3px 8px;
  border-radius: var(--radius-sm);
  white-space: nowrap;
}

.item-badge-green {
  background-color: var(--color-accent-green-soft);
  color: var(--color-accent-green);
}

.privacy-outro {
  display: flex;
  align-items: center;
  gap: 16px;
  padding-bottom: 48px;
  flex-wrap: wrap;
}

.outro-stamp {
  display: inline-block;
  border: 3px solid var(--color-accent-green);
  color: var(--color-accent-green);
  font-size: 0.75rem;
  font-weight: 900;
  letter-spacing: 0.12em;
  padding: 6px 14px;
  border-radius: 4px;
  transform: rotate(-2deg);
  opacity: 0.85;
}

.privacy-outro p {
  color: var(--color-text-muted);
  font-size: 0.9rem;
  margin: 0;
}

@media (max-width: 600px) {
  .privacy-item {
    flex-wrap: wrap;
  }
  .item-badge {
    order: -1;
  }
}
</style>
