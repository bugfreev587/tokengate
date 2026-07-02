<template>
  <div class="flex min-h-screen items-center justify-center bg-gray-50 px-6 text-gray-600 dark:bg-dark-950 dark:text-dark-300">
    <div class="text-center">
      <div class="mx-auto mb-4 h-10 w-10 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"></div>
      <p class="text-sm font-medium">{{ t('common.loading') }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { authAPI } from '@/api'
import { normalizeSessionBridgeRedirect } from '@/utils/sessionBridge'

const route = useRoute()
const { t } = useI18n()

async function prepareSessionCookies(): Promise<void> {
  if (authAPI.getAuthToken()) {
    try {
      await authAPI.prepareSessionCookies()
      return
    } catch {
      // Fall through to refresh-token based recovery below.
    }
  }

  await authAPI.refreshToken()
}

onMounted(async () => {
  const redirectTo = normalizeSessionBridgeRedirect(route.query.redirect)

  try {
    await prepareSessionCookies()
  } catch {
    // No reusable legacy session on this origin. Return to login normally.
  } finally {
    window.location.replace(redirectTo)
  }
})
</script>
