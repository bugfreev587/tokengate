<template>
  <BaseDialog
    :show="show"
    :title="t('admin.accounts.testAccountConnection')"
    width="normal"
    @close="handleClose"
  >
    <div class="space-y-4">
      <!-- Account Info Card -->
      <div
        v-if="account"
        class="flex items-center justify-between rounded-xl border border-gray-200 bg-gradient-to-r from-gray-50 to-gray-100 p-3 dark:border-dark-500 dark:from-dark-700 dark:to-dark-600"
      >
        <div class="flex items-center gap-3">
          <div
            class="flex h-10 w-10 items-center justify-center rounded-lg bg-gradient-to-br from-primary-500 to-primary-600"
          >
            <Icon name="play" size="md" class="text-white" :stroke-width="2" />
          </div>
          <div>
            <div class="font-semibold text-gray-900 dark:text-gray-100">{{ account.name }}</div>
            <div class="flex items-center gap-1.5 text-xs text-gray-500 dark:text-gray-400">
              <span
                class="rounded bg-gray-200 px-1.5 py-0.5 text-[10px] font-medium uppercase dark:bg-dark-500"
              >
                {{ account.type }}
              </span>
              <span>{{ t('admin.accounts.account') }}</span>
            </div>
          </div>
        </div>
        <span
          :class="[
            'rounded-full px-2.5 py-1 text-xs font-semibold',
            account.status === 'active'
              ? 'bg-green-100 text-green-700 dark:bg-green-500/20 dark:text-green-400'
              : 'bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-400'
          ]"
        >
          {{ account.status }}
        </span>
      </div>

      <div class="space-y-1.5">
        <div class="flex items-center justify-between gap-3">
          <label class="text-sm font-medium text-gray-700 dark:text-gray-300">
            {{ t('admin.accounts.selectTestModel') }}
          </label>
          <button
            v-if="supportsModelRefresh"
            type="button"
            class="rounded-lg p-1.5 text-cyan-600 transition-colors hover:bg-cyan-50 disabled:opacity-50 dark:hover:bg-cyan-900/20"
            :title="t('admin.accounts.refreshModels')"
            :aria-label="t('admin.accounts.refreshModels')"
            :disabled="loadingModels || refreshingModels || status === 'connecting'"
            @click="refreshAvailableModels"
          >
            <Icon name="refresh" size="sm" :class="{ 'animate-spin': refreshingModels }" />
          </button>
        </div>
        <Select
          v-model="selectedModelId"
          :options="availableModels"
          :disabled="loadingModels || status === 'connecting'"
          value-key="id"
          label-key="display_name"
          :placeholder="loadingModels ? t('common.loading') + '...' : t('admin.accounts.selectTestModel')"
        />
      </div>

      <div v-if="isOpenAIAccount" class="space-y-1.5">
        <label class="text-sm font-medium text-gray-700 dark:text-gray-300">
          {{ t('admin.accounts.openai.testMode') }}
        </label>
        <Select
          v-model="testMode"
          :options="openAITestModeOptions"
          :disabled="status === 'connecting'"
        />
      </div>

      <div v-if="supportsImageTest" class="space-y-1.5">
        <TextArea
          v-model="testPrompt"
          :label="t('admin.accounts.imagePromptLabel')"
          :placeholder="t('admin.accounts.imagePromptPlaceholder')"
          :hint="t('admin.accounts.imageTestHint')"
          :disabled="status === 'connecting'"
          rows="3"
        />
      </div>

      <!-- Terminal Output -->
      <div class="group relative">
        <div
          ref="terminalRef"
          class="max-h-[240px] min-h-[120px] overflow-y-auto rounded-xl border border-gray-700 bg-gray-900 p-4 font-mono text-sm dark:border-gray-800 dark:bg-black"
        >
          <!-- Status Line -->
          <div v-if="status === 'idle'" class="flex items-center gap-2 text-gray-500">
            <Icon name="play" size="sm" :stroke-width="2" />
            <span>{{ t('admin.accounts.readyToTest') }}</span>
          </div>
          <div v-else-if="status === 'connecting'" class="flex items-center gap-2 text-yellow-400">
            <Icon name="refresh" size="sm" class="animate-spin" :stroke-width="2" />
            <span>{{ t('admin.accounts.connectingToApi') }}</span>
          </div>

          <!-- Output Lines -->
          <div v-for="(line, index) in outputLines" :key="index" :class="line.class">
            {{ line.text }}
          </div>

          <!-- Streaming Content -->
          <div v-if="streamingContent" class="text-green-400">
            {{ streamingContent }}<span class="animate-pulse">_</span>
          </div>

          <!-- Result Status -->
          <div
            v-if="status === 'success'"
            class="mt-3 flex items-center gap-2 border-t border-gray-700 pt-3 text-green-400"
          >
            <Icon name="check" size="sm" :stroke-width="2" />
            <span>{{ t('admin.accounts.testCompleted') }}</span>
          </div>
          <div
            v-else-if="status === 'error'"
            class="mt-3 flex items-center gap-2 border-t border-gray-700 pt-3 text-red-400"
          >
            <Icon name="x" size="sm" :stroke-width="2" />
            <span>{{ errorMessage }}</span>
          </div>
        </div>

        <!-- Copy Button -->
        <button
          v-if="outputLines.length > 0"
          @click="copyOutput"
          class="absolute right-2 top-2 rounded-lg bg-gray-800/80 p-1.5 text-gray-400 opacity-0 transition-all hover:bg-gray-700 hover:text-white group-hover:opacity-100"
          :title="t('admin.accounts.copyOutput')"
        >
          <Icon name="link" size="sm" :stroke-width="2" />
        </button>
      </div>

      <div v-if="generatedImages.length > 0" class="space-y-2">
        <div class="text-xs font-medium text-gray-600 dark:text-gray-300">
          {{ t('admin.accounts.imagePreview') }}
        </div>
        <div class="flex flex-wrap justify-center gap-3">
          <div
            v-for="(image, index) in generatedImages"
            :key="`${image.url}-${index}`"
            class="group/img relative cursor-pointer overflow-hidden rounded-xl border border-gray-200 bg-white shadow-sm transition hover:border-primary-300 hover:shadow-md dark:border-dark-500 dark:bg-dark-700"
            @click="previewImageUrl = image.url"
          >
            <img :src="image.url" :alt="`test-image-${index + 1}`" class="max-h-[360px] w-full object-contain" />
            <div class="absolute inset-0 flex items-center justify-center bg-black/0 transition-colors group-hover/img:bg-black/20">
              <Icon name="eye" size="lg" class="text-white opacity-0 drop-shadow-lg transition-opacity group-hover/img:opacity-100" :stroke-width="2" />
            </div>
            <div class="border-t border-gray-100 px-3 py-1.5 text-xs text-gray-500 dark:border-dark-500 dark:text-gray-300">
              {{ image.mimeType || 'image/*' }}
            </div>
          </div>
        </div>
      </div>

      <!-- Image Lightbox -->
      <Teleport to="body">
        <Transition name="fade">
          <div
            v-if="previewImageUrl"
            class="fixed inset-0 z-[100] flex items-center justify-center bg-black/80 p-4"
            @click.self="previewImageUrl = ''"
          >
            <button
              class="absolute right-4 top-4 rounded-full bg-black/50 p-2 text-white transition-colors hover:bg-black/70"
              @click="previewImageUrl = ''"
            >
              <Icon name="x" size="lg" :stroke-width="2" />
            </button>
            <img
              :src="previewImageUrl"
              alt="preview"
              class="max-h-[90vh] max-w-[90vw] rounded-lg object-contain shadow-2xl"
            />
          </div>
        </Transition>
      </Teleport>

      <!-- Test Info -->
      <div class="flex items-center justify-between px-1 text-xs text-gray-500 dark:text-gray-400">
        <div class="flex items-center gap-3">
          <span class="flex items-center gap-1">
            <Icon name="grid" size="sm" :stroke-width="2" />
            {{ t('admin.accounts.testModel') }}
          </span>
        </div>
        <span class="flex items-center gap-1">
          <Icon name="chat" size="sm" :stroke-width="2" />
          {{
            supportsImageTest
              ? t('admin.accounts.imageTestMode')
              : t('admin.accounts.testPrompt')
          }}
        </span>
      </div>
    </div>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button
          @click="handleClose"
          class="rounded-lg bg-gray-100 px-4 py-2 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-200 dark:bg-dark-600 dark:text-gray-300 dark:hover:bg-dark-500"
        >
          {{ t('common.close') }}
        </button>
        <button
          @click="startTest"
          :disabled="status === 'connecting' || !selectedModelId"
          :class="[
            'flex items-center gap-2 rounded-lg px-4 py-2 text-sm font-medium transition-all',
            status === 'connecting' || !selectedModelId
              ? 'cursor-not-allowed bg-primary-400 text-white'
              : status === 'success'
                ? 'bg-green-500 text-white hover:bg-green-600'
                : status === 'error'
                  ? 'bg-orange-500 text-white hover:bg-orange-600'
                  : 'bg-primary-500 text-white hover:bg-primary-600'
          ]"
        >
          <Icon
            v-if="status === 'connecting'"
            name="refresh"
            size="sm"
            class="animate-spin"
            :stroke-width="2"
          />
          <Icon v-else-if="status === 'idle'" name="play" size="sm" :stroke-width="2" />
          <Icon v-else name="refresh" size="sm" :stroke-width="2" />
          <span>
            {{
              status === 'connecting'
                ? t('admin.accounts.testing')
                : status === 'idle'
                  ? t('admin.accounts.startTest')
                  : t('admin.accounts.retry')
            }}
          </span>
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import TextArea from '@/components/common/TextArea.vue'
import { Icon } from '@/components/icons'
import { useClipboard } from '@/composables/useClipboard'
import { adminAPI } from '@/api/admin'
import { getConnectedAccountModels, refreshConnectedAccountModels } from '@/api/user'
import { useAppStore } from '@/stores/app'
import type { ClaudeModel } from '@/types'

const { t } = useI18n()
const { copyToClipboard } = useClipboard()
const appStore = useAppStore()

interface OutputLine {
  text: string
  class: string
}

interface PreviewImage {
  url: string
  mimeType?: string
}

type AccountTestScope = 'admin' | 'user'

interface TestableAccount {
  id: number
  name: string
  platform: string
  type: string
  status: string
}

const props = withDefaults(defineProps<{
  show: boolean
  account: TestableAccount | null
  scope?: AccountTestScope
}>(), {
  scope: 'admin'
})

const emit = defineEmits<{
  (e: 'close'): void
}>()

const terminalRef = ref<HTMLElement | null>(null)
const status = ref<'idle' | 'connecting' | 'success' | 'error'>('idle')
const outputLines = ref<OutputLine[]>([])
const streamingContent = ref('')
const errorMessage = ref('')
const availableModels = ref<ClaudeModel[]>([])
const selectedModelId = ref('')
const testPrompt = ref('')
const loadingModels = ref(false)
const refreshingModels = ref(false)
let abortController: AbortController | null = null
const generatedImages = ref<PreviewImage[]>([])
const testMode = ref<'default' | 'compact'>('default')
const isOpenAIAccount = computed(() => props.account?.platform === 'openai')
const openAITestModeOptions = computed(() => [
  { value: 'default', label: t('admin.accounts.openai.testModeDefault') },
  { value: 'compact', label: t('admin.accounts.openai.testModeCompact') }
])
const previewImageUrl = ref('')
const prioritizedGeminiModels = ['gemini-3.1-flash-image', 'gemini-2.5-flash-image', 'gemini-2.5-flash', 'gemini-2.5-pro', 'gemini-3-flash-preview', 'gemini-3-pro-preview', 'gemini-2.0-flash']
const TEST_ALL_MODELS_ID = '__all_models__'
const activeTestModelId = ref('')

const isTestAllModelsOption = (modelID: string) => modelID === TEST_ALL_MODELS_ID

const createTestAllModelsOption = (): ClaudeModel => ({
  id: TEST_ALL_MODELS_ID,
  type: 'test-option',
  display_name: t('admin.accounts.testAllModelsConnection'),
  created_at: ''
})

const concreteAvailableModels = computed(() => availableModels.value.filter((model) => !isTestAllModelsOption(model.id)))

const supportsGeminiImageModel = (modelID: string) => {
  const normalizedModelID = modelID.toLowerCase()
  if (!normalizedModelID.startsWith('gemini-') || !normalizedModelID.includes('-image')) return false

  return props.account?.platform === 'gemini' || (props.account?.platform === 'antigravity' && props.account?.type === 'apikey')
}

const supportsOpenAIImageModel = (modelID: string) => {
  const normalizedModelID = modelID.toLowerCase()
  if (!normalizedModelID.startsWith('gpt-image-')) return false
  return props.account?.platform === 'openai'
}

const supportsImageModel = (modelID: string) => supportsGeminiImageModel(modelID) || supportsOpenAIImageModel(modelID)
const supportsImageTest = computed(() => !isTestAllModelsOption(selectedModelId.value) && supportsImageModel(selectedModelId.value))
const activeRequestSupportsImageTest = computed(() => {
  const modelID = activeTestModelId.value || selectedModelId.value
  return !isTestAllModelsOption(modelID) && supportsImageModel(modelID)
})
const supportsModelRefresh = computed(() => {
  return props.account?.platform === 'anthropic' && ['oauth', 'setup-token', 'apikey'].includes(props.account?.type || '')
})
const isStandaloneBuild = import.meta.env.VITE_BUILD_TARGET === 'standalone'

const normalizeApiBase = (value?: string | null) => {
  const raw = (value || '').trim()
  if (!raw) return ''

  try {
    const url = new URL(raw, window.location.origin)
    url.pathname = url.pathname.replace(/\/+$/, '')
    return `${url.origin}${url.pathname}`
  } catch {
    return raw.replace(/\/+$/, '')
  }
}

const resolvedApiBase = computed(() => {
  const candidates = [
    import.meta.env.VITE_API_BASE_URL as string | undefined,
    appStore.apiBaseUrl,
    '/api/v1'
  ]

  for (const candidate of candidates) {
    const normalized = normalizeApiBase(candidate)
    if (!normalized) continue
    if (isStandaloneBuild) {
      try {
        const url = new URL(normalized, window.location.origin)
        if (url.origin === window.location.origin) continue
      } catch {
        if (normalized.startsWith('/')) continue
      }
    }
    return normalized
  }

  return '/api/v1'
})

const accountEndpointPrefix = computed(() => (
  props.scope === 'user' ? 'user/accounts' : 'admin/accounts'
))

const sortTestModels = (models: ClaudeModel[]) => {
  const priorityMap = new Map(prioritizedGeminiModels.map((id, index) => [id, index]))

  return [...models].sort((a, b) => {
    const aPriority = priorityMap.get(a.id) ?? Number.MAX_SAFE_INTEGER
    const bPriority = priorityMap.get(b.id) ?? Number.MAX_SAFE_INTEGER
    if (aPriority !== bPriority) return aPriority - bPriority
    return 0
  })
}

// Load available models when modal opens
watch(
  () => props.show,
  async (newVal) => {
    if (newVal && props.account) {
      testPrompt.value = ''
      testMode.value = 'default'
      resetState()
      await loadAvailableModels()
    } else {
      abortStream()
    }
  }
)

watch(selectedModelId, () => {
  if (supportsImageTest.value && !testPrompt.value.trim()) {
    testPrompt.value = t('admin.accounts.imagePromptDefault')
  }
})

const applyAvailableModels = (models: ClaudeModel[], preferredModelId = '') => {
  if (!props.account) return

  const sortedModels = props.account.platform === 'gemini' || props.account.platform === 'antigravity'
    ? sortTestModels(models)
    : models

  availableModels.value = sortedModels.length > 0
    ? [createTestAllModelsOption(), ...sortedModels]
    : []

  if (availableModels.value.length === 0) {
    selectedModelId.value = ''
    return
  }

  const preferred = preferredModelId
    ? concreteAvailableModels.value.find((model) => model.id === preferredModelId)
    : undefined
  if (preferred) {
    selectedModelId.value = preferred.id
    return
  }

  selectedModelId.value = TEST_ALL_MODELS_ID
}

const loadAvailableModels = async () => {
  if (!props.account) return

  loadingModels.value = true
  selectedModelId.value = '' // Reset selection before loading
  try {
    const models = props.scope === 'user'
      ? await getConnectedAccountModels(props.account.id)
      : await adminAPI.accounts.getAvailableModels(props.account.id)
    applyAvailableModels(models)
  } catch (error) {
    console.error('Failed to load available models:', error)
    // Fallback to empty list
    availableModels.value = []
    selectedModelId.value = ''
  } finally {
    loadingModels.value = false
  }
}

const refreshAvailableModels = async () => {
  if (!props.account) return

  refreshingModels.value = true
  try {
    const models = props.scope === 'user'
      ? await refreshConnectedAccountModels(props.account.id)
      : await adminAPI.accounts.refreshModels(props.account.id)
    applyAvailableModels(models, selectedModelId.value)
    appStore.showSuccess(t('admin.accounts.modelsRefreshed'))
  } catch (error: any) {
    console.error('Failed to refresh available models:', error)
    appStore.showError(error?.message || t('admin.accounts.refreshModelsFailed'))
  } finally {
    refreshingModels.value = false
  }
}

const resetState = () => {
  status.value = 'idle'
  outputLines.value = []
  streamingContent.value = ''
  errorMessage.value = ''
  generatedImages.value = []
  previewImageUrl.value = ''
  activeTestModelId.value = ''
}

const handleClose = () => {
  abortStream()
  emit('close')
}

const abortStream = () => {
  if (abortController) {
    abortController.abort()
    abortController = null
  }
}

const addLine = (text: string, className: string = 'text-gray-300') => {
  outputLines.value.push({ text, class: className })
  scrollToBottom()
}

const scrollToBottom = async () => {
  await nextTick()
  if (terminalRef.value) {
    terminalRef.value.scrollTop = terminalRef.value.scrollHeight
  }
}

interface AccountTestEvent {
  type: string
  text?: string
  model?: string
  success?: boolean
  error?: string
  image_url?: string
  mime_type?: string
}

const isAbortError = (error: unknown) => error instanceof DOMException && error.name === 'AbortError'

const getModelLabel = (model: ClaudeModel) => model.display_name || model.id

const flushStreamingContent = () => {
  if (streamingContent.value) {
    addLine(streamingContent.value, 'text-green-300')
    streamingContent.value = ''
  }
}

const currentTestFailed = () => status.value === 'error'

const getPromptForModel = (modelID: string) => {
  if (!supportsImageModel(modelID)) return ''
  return testPrompt.value.trim() || t('admin.accounts.imagePromptDefault')
}

const streamAccountTest = async (modelID: string) => {
  if (!props.account) return

  activeTestModelId.value = modelID
  streamingContent.value = ''

  const url = `${resolvedApiBase.value}/${accountEndpointPrefix.value}/${props.account.id}/test`
  const response = await fetch(url, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${localStorage.getItem('auth_token')}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      model_id: modelID,
      prompt: getPromptForModel(modelID),
      mode: isOpenAIAccount.value ? testMode.value : 'default'
    }),
    signal: abortController?.signal
  })

  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`)
  }

  const reader = response.body?.getReader()
  if (!reader) {
    throw new Error('No response body')
  }

  const decoder = new TextDecoder()
  let buffer = ''

  while (true) {
    const { done, value } = await reader.read()
    if (done) break

    buffer += decoder.decode(value, { stream: true })
    const lines = buffer.split('\n')
    buffer = lines.pop() || ''

    for (const line of lines) {
      if (!line.startsWith('data: ')) continue

      const jsonStr = line.slice(6).trim()
      if (!jsonStr) continue

      try {
        const event = JSON.parse(jsonStr) as AccountTestEvent
        handleEvent(event)
      } catch (e) {
        console.error('Failed to parse SSE event:', e)
      }
    }
  }
}

const startAllModelsTest = async () => {
  const models = concreteAvailableModels.value
  if (models.length === 0) {
    throw new Error(t('admin.accounts.noModelsAvailableForTest'))
  }

  addLine(t('admin.accounts.testingAllModels', { count: models.length }), 'text-cyan-400')

  let passed = 0
  let failed = 0

  for (const [index, model] of models.entries()) {
    const modelLabel = getModelLabel(model)
    if (index > 0) {
      addLine('', 'text-gray-300')
    }
    status.value = 'connecting'
    addLine(t('admin.accounts.testingModel', { model: modelLabel }), 'text-cyan-400')

    try {
      await streamAccountTest(model.id)
      flushStreamingContent()

      if (currentTestFailed()) {
        failed += 1
        addLine(
          t('admin.accounts.testAllModelsModelFailed', {
            model: modelLabel,
            error: errorMessage.value || t('admin.accounts.testFailed')
          }),
          'text-red-400'
        )
        errorMessage.value = ''
        continue
      }

      passed += 1
      addLine(t('admin.accounts.testAllModelsModelPassed', { model: modelLabel }), 'text-green-400')
    } catch (error: unknown) {
      if (isAbortError(error)) throw error

      failed += 1
      const msg = error instanceof Error ? error.message : 'Unknown error'
      addLine(
        t('admin.accounts.testAllModelsModelFailed', {
          model: modelLabel,
          error: msg
        }),
        'text-red-400'
      )
      errorMessage.value = ''
    }
  }

  addLine('', 'text-gray-300')
  if (failed > 0) {
    status.value = 'error'
    errorMessage.value = t('admin.accounts.testAllModelsFailedSummary', {
      passed,
      total: models.length,
      failed
    })
    addLine(errorMessage.value, 'text-red-400')
    return
  }

  status.value = 'success'
  addLine(t('admin.accounts.testAllModelsSummary', { passed, total: models.length }), 'text-green-400')
}

const startTest = async () => {
  if (!props.account || !selectedModelId.value) return

  resetState()
  status.value = 'connecting'
  addLine(t('admin.accounts.startingTestForAccount', { name: props.account.name }), 'text-blue-400')
  addLine(t('admin.accounts.testAccountTypeLabel', { type: props.account.type }), 'text-gray-400')
  addLine('', 'text-gray-300')

  abortStream()

  abortController = new AbortController()

  try {
    if (isTestAllModelsOption(selectedModelId.value)) {
      await startAllModelsTest()
    } else {
      await streamAccountTest(selectedModelId.value)
    }
  } catch (error: unknown) {
    if (isAbortError(error)) {
      status.value = 'idle'
      return
    }
    status.value = 'error'
    const msg = error instanceof Error ? error.message : 'Unknown error'
    errorMessage.value = msg
    addLine(`Error: ${msg}`, 'text-red-400')
  } finally {
    activeTestModelId.value = ''
  }
}

const handleEvent = (event: AccountTestEvent) => {
  switch (event.type) {
    case 'test_start':
      addLine(t('admin.accounts.connectedToApi'), 'text-green-400')
      if (event.model) {
        addLine(t('admin.accounts.usingModel', { model: event.model }), 'text-cyan-400')
      }
      addLine(
        activeRequestSupportsImageTest.value
            ? t('admin.accounts.sendingImageRequest')
            : t('admin.accounts.sendingTestMessage'),
        'text-gray-400'
      )
      addLine('', 'text-gray-300')
      addLine(t('admin.accounts.response'), 'text-yellow-400')
      break

    case 'content':
      if (event.text) {
        streamingContent.value += event.text
        scrollToBottom()
      }
      break

    case 'image':
      if (event.image_url) {
        generatedImages.value.push({
          url: event.image_url,
          mimeType: event.mime_type
        })
        addLine(t('admin.accounts.imageReceived', { count: generatedImages.value.length }), 'text-purple-300')
      }
      break

    case 'test_complete':
      // Move streaming content to output lines
      if (streamingContent.value) {
        addLine(streamingContent.value, 'text-green-300')
        streamingContent.value = ''
      }
      if (event.success) {
        status.value = 'success'
      } else {
        status.value = 'error'
        errorMessage.value = event.error || 'Test failed'
      }
      break

    case 'error':
      status.value = 'error'
      errorMessage.value = event.error || 'Unknown error'
      if (streamingContent.value) {
        addLine(streamingContent.value, 'text-green-300')
        streamingContent.value = ''
      }
      break
  }
}

const copyOutput = () => {
  const text = outputLines.value.map((l) => l.text).join('\n')
  copyToClipboard(text, t('admin.accounts.outputCopied'))
}
</script>

<style>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
