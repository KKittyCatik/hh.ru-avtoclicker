export type AccountPreferences = {
  min_salary: number;
  schedule: string;
  exclude_agency: boolean;
  blacklist_ids: string[];
  blacklist_names: string[];
};

export type AccountToken = {
  access_token: string;
  refresh_token: string;
  expires_at: string;
};

export type Account = {
  id: string;
  name: string;
  token: AccountToken;
  resume_ids: string[];
  search_urls: string[];
  preferences: AccountPreferences;
  needs_reauth: boolean;
};
