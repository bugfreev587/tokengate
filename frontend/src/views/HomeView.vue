<template>
  <div v-if="homeContent" class="min-h-[100dvh]">
    <iframe
      v-if="isHomeContentUrl"
      :src="homeContent.trim()"
      class="h-[100dvh] w-full border-0"
      allowfullscreen
    ></iframe>
    <div v-else v-html="homeContent"></div>
  </div>

  <div
    v-else
    class="relative min-h-[100dvh] overflow-hidden bg-stone-50 text-gray-950 dark:bg-dark-950 dark:text-white"
  >
    <div
      aria-hidden="true"
      class="pointer-events-none absolute inset-x-0 top-0 h-[42rem] bg-[radial-gradient(circle_at_80%_15%,rgba(34,122,99,0.10),transparent_36%),radial-gradient(circle_at_18%_8%,rgba(120,113,108,0.08),transparent_32%)] dark:bg-[radial-gradient(circle_at_80%_15%,rgba(34,122,99,0.12),transparent_34%),radial-gradient(circle_at_18%_8%,rgba(120,113,108,0.05),transparent_30%)]"
    ></div>

    <div class="relative">
      <HomePublicHeader
        :site-name="siteName"
        :site-logo="siteLogo"
        :doc-url="docUrl"
        :doc-url-external="isDocUrlExternal"
        :is-dark="isDark"
        :is-authenticated="isAuthenticated"
        :dashboard-path="dashboardPath"
        :user-initial="userInitial"
        @toggle-theme="toggleTheme"
      />

      <main>
        <HomeCapacityHero
          :site-subtitle="siteSubtitle"
          :doc-url="docUrl"
          :doc-url-external="isDocUrlExternal"
        />
        <HomeSharedCapabilities />
        <HomeHowItWorks />
        <HomeClosingCta />
      </main>

      <HomePublicFooter
        :site-name="siteName"
        :current-year="currentYear"
        :doc-url="docUrl"
        :doc-url-external="isDocUrlExternal"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

import HomeCapacityHero from '@/components/home/HomeCapacityHero.vue'
import HomeClosingCta from '@/components/home/HomeClosingCta.vue'
import HomeHowItWorks from '@/components/home/HomeHowItWorks.vue'
import HomePublicFooter from '@/components/home/HomePublicFooter.vue'
import HomePublicHeader from '@/components/home/HomePublicHeader.vue'
import HomeSharedCapabilities from '@/components/home/HomeSharedCapabilities.vue'
import { useAppStore, useAuthStore } from '@/stores'

const authStore = useAuthStore()
const appStore = useAppStore()

const siteName = computed(
  () => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'TokenGate'
)
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || '')

const defaultDocsUrl = '/docs'
const normalizeDocsUrl = (url?: string) => {
  const trimmed = url?.trim()
  if (!trimmed || trimmed.includes('TOKENGATE_QUICKSTART.md')) {
    return defaultDocsUrl
  }
  return trimmed
}

const docUrl = computed(() =>
  normalizeDocsUrl(appStore.cachedPublicSettings?.doc_url || appStore.docUrl)
)
const isDocUrlExternal = computed(() => /^https?:\/\//.test(docUrl.value))
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')
const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

const isDark = ref(document.documentElement.classList.contains('dark'))
const isAuthenticated = computed(() => authStore.isAuthenticated)
const dashboardPath = computed(() => (authStore.isAdmin ? '/admin/dashboard' : '/dashboard'))
const userInitial = computed(() => {
  const identity = authStore.user?.username || authStore.user?.email || ''
  return identity.charAt(0).toUpperCase()
})
const currentYear = computed(() => new Date().getFullYear())

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

function initTheme() {
  const savedTheme = localStorage.getItem('theme')
  if (
    savedTheme === 'dark' ||
    (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)
  ) {
    isDark.value = true
    document.documentElement.classList.add('dark')
  }
}

onMounted(() => {
  initTheme()
  authStore.checkAuth()

  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
})
</script>
