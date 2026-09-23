"use client";
import { FaBullseye } from "react-icons/fa";
import { PredictionContest } from "../types";
import PredictionListItem from "./PredictionListItem";
import CreatePredictionButton from "./CreatePredictionButton";
import StateMessage from "@/app/components/ui/StateMessage";
import { useUser } from "@/app/components/UserProvider";

type Props = {
  predictions: PredictionContest[];
};

const PredictionList = ({ predictions }: Props) => {
  const { user } = useUser();
  const isAdmin = user?.roles.includes("admin");

  return (
    <>
      {isAdmin && (
        <div className="mb-6 flex flex-wrap items-center gap-3 rounded-card border border-line bg-surface p-3">
          <CreatePredictionButton />
        </div>
      )}

      {predictions.length === 0 ? (
        <StateMessage
          icon={<FaBullseye className="h-5 w-5" />}
          title="Még nincs tippverseny"
          description="Amint elindul az első tippverseny, itt fogod látni."
        />
      ) : (
        <div className="flex flex-col gap-4">
          {predictions.map((prediction) => (
            <PredictionListItem
              key={prediction.id}
              prediction={prediction}
              isAdmin={isAdmin}
            />
          ))}
        </div>
      )}
    </>
  );
};

export default PredictionList;
