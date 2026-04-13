import { crud } from './_crud';
import type { Person } from '@/types';
export const people = crud<Person>('/people');
