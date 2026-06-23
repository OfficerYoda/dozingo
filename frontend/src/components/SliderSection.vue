<template>
    <Splide :options="splideOptions" class="slider-splide" aria-label="slider">
      <SplideSlide v-for="(item, index) in items" :key="index">
        <div class="slider-slide-inner">
          <slot name="slide" :item="item" :index="index" />
        </div>
      </SplideSlide>
    </Splide>
</template>

<script setup lang="ts" generic="T">
import { computed } from 'vue'
import { Splide, SplideSlide } from '@splidejs/vue-splide'
import '@splidejs/vue-splide/css/core'

const props = withDefaults(defineProps<{
  items: T[]
  /** How many slides are visible at once on desktop (≥ 900px). Default: 3 */
  perPage?: number
  /** How many slides are visible at once on tablet (< 900px). Default: 2 */
  perPageMd?: number
  /** How many slides are visible at once on mobile (< 600px). Default: 1 */
  perPageSm?: number
  /** Gap between slides, any CSS value. Default: '1rem' */
  gap?: string
  /** Whether autoplay is enabled. Default: false */
  autoplay?: boolean
  /** Autoplay interval in ms. Default: 4000 */
  interval?: number
  /** Slider type: 'slide' | 'loop' | 'fade'. Default: 'slide' */
  type?: 'slide' | 'loop' | 'fade'
}>(), {
  perPage: 3,
  perPageMd: 2,
  perPageSm: 1,
  gap: '1rem',
  autoplay: false,
  interval: 4000,
  type: 'slide',
})

const splideOptions = computed(() => ({
  type: props.type,
  perPage: props.perPage,
  perMove: 1,
  gap: props.gap,
  autoplay: props.autoplay,
  interval: props.interval,
  pauseOnHover: true,
  rewind: false,
  pagination: true,
  arrows: true,
  breakpoints: {
    900: {
      perPage: props.perPageMd,
    },
    600: {
      perPage: props.perPageSm,
    },
  },
}))
</script>

<style>
.slider-splide .splide__track {
  padding-bottom: 1rem; 
}

.slider-splide .splide__pagination {
  bottom: 0;
  gap: 0.4rem;
}

.slider-splide .splide__pagination__page {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--color-border-strong);
  border: none;
  cursor: pointer;
  transition: background var(--transition-fast), transform var(--transition-fast);
  padding: 0;
  margin: 0;
}

.slider-splide .splide__pagination__page.is-active {
  background: var(--color-primary-500);
  transform: scale(1.25);
}

.slider-splide .splide__arrow {
  position: absolute;
  top: calc(50% - 1.25rem);
  transform: translateY(-50%);
  z-index: 1;
  width: 2.25rem;
  height: 2.25rem;
  border-radius: 50%;
  background: var(--color-bg-surface);
  border: 1px solid var(--color-border);
  box-shadow: var(--shadow-sm);
  color: var(--color-text-primary);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background var(--transition-fast), box-shadow var(--transition-fast);
}

.slider-splide .splide__arrow:hover:not(:disabled) {
  background: var(--color-bg-hover);
  box-shadow: var(--shadow-md);
}

.slider-splide .splide__arrow:disabled {
  opacity: 0.3;
  cursor: default;
  pointer-events: none;
}

.slider-splide .splide__arrow svg {
  width: 1rem;
  height: 1rem;
  fill: currentColor;
}

.slider-splide .splide__arrow--prev svg {
  transform: scaleX(-1);
}

.slider-splide .splide__arrow--prev {
  left: -1rem;
}

.slider-splide .splide__arrow--next {
  right: -1rem;
}

.slider-slide-inner {
  height: 100%;
}
</style>
