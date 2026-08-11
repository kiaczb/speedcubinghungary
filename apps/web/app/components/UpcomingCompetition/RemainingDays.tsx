import { RiTimeFill } from "react-icons/ri";

type Props = {
  days: number;
};

const RemainingDays = ({ days }: Props) => {
  const isToday = days <= 0;

  return (
    <div className="mt-7 flex items-center justify-center gap-4 rounded-card border border-white/10 bg-white/5 p-4">
      <RiTimeFill className="h-6 w-6 shrink-0 text-white/30" />

      {isToday ? (
        <p className="text-lg font-bold text-white">Már zajlik</p>
      ) : (
        <div className="flex items-baseline gap-2">
          <p className="tabular font-mono text-3xl font-bold leading-none text-white sm:text-4xl">
            {days}
          </p>
          <span className="font-mono text-xs uppercase tracking-[0.18em] text-white/40">
            nap múlva
          </span>
        </div>
      )}
    </div>
  );
};

export default RemainingDays;
