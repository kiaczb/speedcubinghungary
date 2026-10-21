export interface UpcomingCompetition {
  id: string;
  name: string;
  startDate: string;
  endDate: string;
  city: string;
  country: string;
  url: string;
  eventIds: string[];
  registrationOpen: string;
  registrationClose: string;
  competitorLimit: number | null;
  acceptedCount: number | null;
}
