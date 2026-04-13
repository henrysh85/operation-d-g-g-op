import { crud } from './_crud';
import type { Jurisdiction } from '@/types';
export const regulatory = crud<Jurisdiction>('/regulatory/jurisdictions');
