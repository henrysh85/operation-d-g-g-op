import { crud } from './_crud';
import type { Client } from '@/types';
export const clients = crud<Client>('/clients');
