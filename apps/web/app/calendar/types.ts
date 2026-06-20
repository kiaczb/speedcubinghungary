// Shape returned by our backend: GET /competitions/upcoming
export interface UpcomingCompetition {
  id: string;
  name: string;
  startDate: string; // YYYY-MM-DD
  endDate: string; // YYYY-MM-DD
  city: string;
  country: string;
  url: string;
}
