import { UpcomingCompetition } from "../types";

export type RegistrationState =
  | { kind: "upcoming"; opensAt: string }
  | { kind: "open"; closesAt: string; spotsLeft: number | null }
  | { kind: "full"; closesAt: string }
  | { kind: "closed" }
  | { kind: "unknown" };

const timestamp = (value: string) => {
  const time = new Date(value).getTime();
  return Number.isNaN(time) ? null : time;
};

export function registrationState(
  competition: UpcomingCompetition,
  now: Date,
): RegistrationState {
  const opens = timestamp(competition.registrationOpen);
  const closes = timestamp(competition.registrationClose);
  if (opens === null || closes === null) return { kind: "unknown" };

  const current = now.getTime();
  if (current < opens) {
    return { kind: "upcoming", opensAt: competition.registrationOpen };
  }
  if (current >= closes) return { kind: "closed" };

  const { competitorLimit, acceptedCount } = competition;
  const spotsLeft =
    competitorLimit !== null && acceptedCount !== null
      ? Math.max(competitorLimit - acceptedCount, 0)
      : null;

  return spotsLeft === 0
    ? { kind: "full", closesAt: competition.registrationClose }
    : { kind: "open", closesAt: competition.registrationClose, spotsLeft };
}
