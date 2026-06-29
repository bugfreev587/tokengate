import { apiClient } from '../client'
import type { BillingMode } from '@/constants/channel'

export interface ModelPricingSnapshot {
  input_price: number | null
  output_price: number | null
  cache_write_price: number | null
  cache_read_price: number | null
  image_output_price: number | null
  per_request_price: number | null
}

export interface GlobalModelPricingOverride extends ModelPricingSnapshot {
  id: number
  model: string
  provider: string
  billing_mode: BillingMode
  created_at: string
  updated_at: string
}

export interface GlobalModelPricingRow extends ModelPricingSnapshot {
  model: string
  provider: string
  billing_mode: BillingMode
  source: 'litellm' | 'fallback' | 'global_override' | string
  fallback: ModelPricingSnapshot | null
  override: GlobalModelPricingOverride | null
  updated_at: string | null
}

export interface GlobalModelPricingListParams {
  page?: number
  page_size?: number
  search?: string
  provider?: string
  source?: string
  billing_mode?: string
}

export interface GlobalModelPricingListResponse {
  items: GlobalModelPricingRow[]
  total: number
  page: number
  page_size: number
  pages: number
}

export interface UpsertGlobalModelPricingOverrideRequest {
  model: string
  provider?: string
  billing_mode?: BillingMode
  input_price?: number | null
  output_price?: number | null
  cache_write_price?: number | null
  cache_read_price?: number | null
  image_output_price?: number | null
  per_request_price?: number | null
}

export async function list(params: GlobalModelPricingListParams = {}): Promise<GlobalModelPricingListResponse> {
  const { data } = await apiClient.get<GlobalModelPricingListResponse>('/admin/model-pricing', { params })
  return data
}

export async function upsertOverride(payload: UpsertGlobalModelPricingOverrideRequest): Promise<GlobalModelPricingOverride> {
  const { data } = await apiClient.post<GlobalModelPricingOverride>('/admin/model-pricing/overrides', payload)
  return data
}

export async function deleteOverride(model: string): Promise<void> {
  await apiClient.delete('/admin/model-pricing/overrides', { params: { model } })
}

const modelPricingAPI = { list, upsertOverride, deleteOverride }
export default modelPricingAPI
