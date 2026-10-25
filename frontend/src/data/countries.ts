// Russian labels for country codes from the countries seed (backend/seeds/development/countries.go).
// The API returns English names; unknown codes fall back to them.
export const COUNTRY_LABELS_RU: Record<string, string> = {
  ENG: 'Англия',
  ESP: 'Испания',
  ITA: 'Италия',
  GER: 'Германия',
  FRA: 'Франция',
  POR: 'Португалия',
  NED: 'Нидерланды',
  TUR: 'Турция',
  NOR: 'Норвегия',
  BEL: 'Бельгия',
  GRE: 'Греция',
  AUT: 'Австрия',
  UKR: 'Украина',
  CZE: 'Чехия',
  SVK: 'Словакия',
  AZE: 'Азербайджан',
}

// Top leagues first in the country filter; the rest follow alphabetically.
export const COUNTRY_PRIORITY = ['ENG', 'ESP', 'ITA', 'GER', 'FRA']
