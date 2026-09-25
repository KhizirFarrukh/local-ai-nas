// Types of the files area, from the API spec (api/openapi.yaml).
import type { components } from '$lib/api/schema';

export type FileItem = components['schemas']['FileItem'];
export type ItemKind = components['schemas']['ItemKind'];
export type SortKey = components['schemas']['ListSort'];
export type SortOrder = components['schemas']['ListOrder'];
export type ItemsResponse = components['schemas']['ItemsResponse'];
