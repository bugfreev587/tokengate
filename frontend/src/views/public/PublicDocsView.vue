<template>
  <ApiPage
    :endpoints="tokenGateApiEndpoints"
    :sidebar-groups="tokenGateApiSidebarGroups"
    :site-name="siteName"
    :site-logo="siteLogo"
  />
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import ApiPage from '@/components/docs/api/ApiPage.vue'
import { useAppStore } from '@/stores/app'
import { tokenGateApiEndpoints, tokenGateApiSidebarGroups } from '@/config/apiReference'

const appStore = useAppStore()

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'TokenGate')
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')

onMounted(() => {
  if (!appStore.cachedPublicSettings) {
    appStore.fetchPublicSettings().catch(() => {})
  }
})
</script>
