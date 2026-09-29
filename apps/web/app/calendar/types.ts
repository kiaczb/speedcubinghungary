// Shape returned by our backend: GET /competitions/upcoming
export interface UpcomingCompetition {
  id: string;
  name: string;
  startDate: string; // YYYY-MM-DD
  endDate: string; // YYYY-MM-DD
  city: string;
  country: string;
  url: string;
  eventIds: string[]; // WCA event ids held at this competition, e.g. "333", "222"
}
