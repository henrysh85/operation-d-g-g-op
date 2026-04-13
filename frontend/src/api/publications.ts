import { crud } from './_crud';
import type { Publication } from '@/types';
export const publications = crud<Publication>('/publications');
