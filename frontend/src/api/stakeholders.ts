import { http } from './client';
import { crud } from './_crud';
import type { Institution, Contact } from '@/types';

export const institutions = crud<Institution>('/stakeholders/institutions');
export const contacts = crud<Contact>('/stakeholders/contacts');

export const stakeholders = {
  institutions,
  contacts,
  async tree(): Promise<Array<{
    region: string;
    countries: Array<{
      countryCode: string;
      institutions: Array<Institution & { contacts: Contact[] }>;
    }>;
  }>> {
    const { data } = await http.get('/stakeholders/tree');
    return data;
  },
};
