// Shared model types — mirror of backend domain objects.
// Kept intentionally permissive: unknown backend extensions are allowed via `extra`.

export type ID = string;
export type ISODate = string;

export type Vertical =
  | 'crypto' | 'banking' | 'securities' | 'insurance'
  | 'payments' | 'fintech' | 'asset-management' | 'other';

export type Region =
  | 'EMEA' | 'Americas' | 'APAC' | 'MENA' | 'Africa' | 'Global';

export type Impact = 'critical' | 'high' | 'medium' | 'low' | 'informational';
export type Status = 'live' | 'active' | 'draft' | 'archived' | 'early' | 'none';

export interface Jurisdiction {
  id: ID;
  name: string;
  code: string;
  region: Region;
  vertical: Vertical;
  status: Status;
  tier: 1 | 2 | 3;
  regulators?: string[];
  updatedAt?: ISODate;
}

export interface Client {
  id: ID;
  name: string;
  shortCode?: string;
  vertical: Vertical;
  tier?: 1 | 2 | 3;
  color?: string;
}

export interface Institution {
  id: ID;
  name: string;
  type: 'regulator' | 'ministry' | 'trade-body' | 'think-tank' | 'ngo' | 'other';
  countryCode: string;
  region: Region;
  website?: string;
}

export interface Contact {
  id: ID;
  name: string;
  title?: string;
  email?: string;
  phone?: string;
  institutionId: ID;
  notes?: string;
  lastInteraction?: ISODate;
}

export interface Person {
  id: ID;
  fullName: string;
  role: string;
  team?: string;
  managerId?: ID | null;
  email?: string;
  office?: string;
  region?: Region;
  startDate?: ISODate;
  isActive: boolean;
}

export interface Activity {
  id: ID;
  type: 'meeting' | 'call' | 'email' | 'research' | 'filing' | 'publication' | 'other';
  title: string;
  summary?: string;
  occurredAt: ISODate;
  personIds: ID[];
  clientIds: ID[];
  jurisdictionIds: ID[];
  tags?: string[];
}

export interface Consultation {
  id: ID;
  title: string;
  regulator: string;
  jurisdictionId: ID;
  vertical: Vertical;
  openedAt: ISODate;
  deadlineAt: ISODate;
  status: 'open' | 'closed' | 'draft' | 'responded';
  impact: Impact;
  url?: string;
}

export interface Template {
  id: ID;
  name: string;
  category: string;
  description?: string;
  owner?: string;
  updatedAt: ISODate;
  downloadUrl?: string;
}

export interface Publication {
  id: ID;
  title: string;
  kind: 'note' | 'briefing' | 'report' | 'alert';
  region: Region;
  publishedAt: ISODate;
  authorIds: ID[];
  summary?: string;
}

export interface Member {
  id: ID;
  legalName: string;
  jurisdictionId: ID;
  status: 'prospect' | 'applied' | 'active' | 'lapsed';
  tier?: 'standard' | 'premium' | 'enterprise';
  joinedAt?: ISODate;
  contactId?: ID;
  riskScore?: number;
}

export interface KPI {
  key: string;
  label: string;
  value: number | string;
  delta?: number;
  unit?: string;
}

export interface User {
  id: ID;
  email: string;
  name: string;
  roles: string[];
}
