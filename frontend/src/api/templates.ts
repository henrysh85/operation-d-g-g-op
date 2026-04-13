import { crud } from './_crud';
import type { Template } from '@/types';
export const templates = crud<Template>('/templates');
