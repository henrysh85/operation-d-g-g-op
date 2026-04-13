import { crud } from './_crud';
import type { Activity } from '@/types';
export const activities = crud<Activity>('/activities');
