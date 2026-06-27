export interface Competition {
  id: string;
  name: string;
  start_date: string; // YYYY-MM-DD
  end_date: string; // YYYY-MM-DD
  registration_open: string; // ISO date string
  registration_close: string; // ISO date string
  announced_at: string; // ISO date string
  cancelled_at: string | null; // ISO date string or null
  results_posted_at: string | null;
  competitor_limit: number;
  venue: string;
  url: string;
  website: string;
  short_name: string;
  short_display_name: string;
  city: string;
  venue_address: string;
  venue_details: string;
  latitude_degrees: number;
  longitude_degrees: number;
  country_iso2: string;
  event_ids: string[];
  time_until_registration: string;
  date_range: string;
  delegates: Person[];
  organizers: Person[];
  class: string;
}

export interface Person {
  id: number;
  wca_id: string;
  name: string;
  gender: string;
  country_iso2: string;
  created_at: string; // ISO date string
  updated_at: string; // ISO date string
  url: string;
  country: Country;
  location: string;
  region_id: number;
  delegate_status: string | null;
  email: string;
  class: string;
  teams: unknown[]; // lehetőség van itt pontosítani, ha tudjuk a struktúrát
  avatar: Avatar;
}

export interface Country {
  id: string;
  name: string;
  continent_id: string;
  iso2: string;
}

export interface Avatar {
  id: number;
  status: string;
  thumbnail_crop_x: number;
  thumbnail_crop_y: number;
  thumbnail_crop_w: number;
  thumbnail_crop_h: number;
  url: string;
  thumb_url: string;
  is_default: boolean;
  can_edit_thumbnail: boolean;
}
