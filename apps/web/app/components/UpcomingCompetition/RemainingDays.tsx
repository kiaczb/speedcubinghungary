import { RiTimeFill } from "react-icons/ri";

type Props = {
  days: number;
};

const RemainingDays = ({ days }: Props) => {
  return (
    <div className="mt-8 flex items-center justify-center gap-4 p-4 rounded-xl bg-linear-to-br from-gray-900/50 to-black/50 border border-white/10 backdrop-blur-sm">
      <RiTimeFill className="text-3xl text-gray-300" />

      <div className="flex items-baseline gap-2">
        <p className="text-3xl md:text-4xl font-bold bg-linear-to-r from-white via-gray-100 to-gray-300 bg-clip-text text-transparent">
          {days}
        </p>
        <span className="text-sm uppercase tracking-widest text-gray-400 font-medium">
          nap
        </span>
      </div>
    </div>
  );
};

export default RemainingDays;
