<template>
  <ApiReferencePage
    :endpoints="tokenGateApiEndpoints"
    :sidebar-groups="tokenGateApiSidebarGroups"
    :site-name="siteName"
    :site-logo="siteLogo"
  />
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import ApiReferencePage from '@/components/docs/api/ApiReferencePage.vue'
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
