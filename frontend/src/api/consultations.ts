import { crud } from './_crud';
import type { Consultation } from '@/types';
export const consultations = crud<Consultation>('/consultations');
