import { describe, expect, it, vi, afterEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'

import Select from '../Select.vue'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  }
})

describe('Select group headers', () => {
  afterEach(() => {
    document.body.innerHTML = ''
  })

  it('keeps a group header visible when search matches an option in that group', async () => {
    const wrapper = mount(Select, {
      attachTo: document.body,
      props: {
        modelValue: null,
        searchable: true,
        options: [
          { value: '__tokengate', label: 'TokenGate Capacity', disabled: true, kind: 'group' },
          { value: 1, label: 'default' },
          { value: '__connected', label: 'My Connected Accounts', disabled: true, kind: 'group' },
          { value: 2, label: 'byo-anthropic-u2-a5' },
        ],
      },
    })

    await wrapper.get('button').trigger('click')
    await nextTick()

    const search = document.body.querySelector<HTMLInputElement>('.select-search-input')
    expect(search).not.toBeNull()

    search!.value = 'anthropic'
    search!.dispatchEvent(new Event('input'))
    await nextTick()

    const dropdownText = document.body.textContent || ''
    expect(dropdownText).toContain('My Connected Accounts')
    expect(dropdownText).toContain('byo-anthropic-u2-a5')
    expect(dropdownText).not.toContain('TokenGate Capacity')
    expect(dropdownText).not.toContain('default')
  })
})
