import { ReactNode } from "react";

type Props = {
  icon: ReactNode;
  title: string;
  description: string;
};

const StateMessage = ({ icon, title, description }: Props) => (
  <div className="flex flex-col items-center justify-center rounded-2xl border border-dashed border-gray-300 bg-white px-6 py-16 text-center">
    <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-gray-50 text-gray-400">
      {icon}
    </div>
    <h3 className="text-base font-semibold text-gray-900">{title}</h3>
    <p className="mt-1 max-w-sm text-sm text-gray-500">{description}</p>
  </div>
);

export default StateMessage;
