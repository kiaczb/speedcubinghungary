"use client";

import { useState } from "react";
import Button from "@/app/components/ui/Button";
import { useCompetitorOptions } from "../lib/useCompetitorOptions";
import { PredictionContest } from "../types";
import PodiumEventForm from "./PodiumEventForm";
import ExtraQuestionField from "./ExtraQuestionField";

type Props = {
  prediction: PredictionContest;
  isOpen: boolean;
};

type PodiumPicks = Record<string, [string, string, string]>;
type ExtraAnswers = Record<string, string>;


const PredictionForm = ({ prediction, isOpen }: Props) => {
  const [podiumPicks, setPodiumPicks] = useState<PodiumPicks>({});
  const [extraAnswers, setExtraAnswers] = useState<ExtraAnswers>({});


  const hasPersonQuestion = prediction.extraQuestions.some(
    (question) => question.type === "person",
  );
  const competitors = useCompetitorOptions(
    hasPersonQuestion
      ? `/competitions/${encodeURIComponent(prediction.competitionId)}/competitors`
      : null,
  );

  const picksFor = (eventId: string): [string, string, string] =>
    podiumPicks[eventId] ?? ["", "", ""];

  const handlePodiumChange = (
    eventId: string,
    place: 0 | 1 | 2,
    name: string,
  ) => {
    setPodiumPicks((prev) => {
      const current = prev[eventId] ?? (["", "", ""] as [string, string, string]);
      const next: [string, string, string] = [...current];
      next[place] = name;
      return { ...prev, [eventId]: next };
    });
  };

  return (
    <form className="mt-8 flex flex-col gap-10">
      <section>
        <h2 className="text-lg font-semibold text-foreground">
          Dobogó-tippek
        </h2>
        <p className="mt-1 text-sm text-muted-foreground">
          Add meg, szerinted ki végez az 1-2-3. helyen az egyes
          versenyszámokban.
        </p>

        <div className="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-2">
          {prediction.podiumEventIds.map((eventId) => (
            <PodiumEventForm
              key={eventId}
              competitionId={prediction.competitionId}
              eventId={eventId}
              picks={picksFor(eventId)}
              onChange={(place, name) =>
                handlePodiumChange(eventId, place, name)
              }
            />
          ))}
        </div>
      </section>

      {prediction.extraQuestions.length > 0 && (
        <section>
          <h2 className="text-lg font-semibold text-foreground">
            Extra kérdések
          </h2>
          <p className="mt-1 text-sm text-muted-foreground">
            Ezekre a kérdésekre is tippelhetsz.
          </p>

          <div className="mt-4 flex flex-col gap-4">
            {prediction.extraQuestions.map((question) => (
              <ExtraQuestionField
                key={question.id}
                question={question}
                competitors={competitors}
                value={extraAnswers[question.id] ?? ""}
                onChange={(value) =>
                  setExtraAnswers((prev) => ({
                    ...prev,
                    [question.id]: value,
                  }))
                }
              />
            ))}
          </div>
        </section>
      )}

      <div className="flex justify-end border-t border-line pt-6">
        <Button type="submit" disabled={!isOpen}>
          {isOpen ? "Tipp leadása" : "A tippelés lezárult"}
        </Button>
      </div>
    </form>
  );
};

export default PredictionForm;
