import { FaArrowRight, FaCalendar, FaRegNewspaper } from "react-icons/fa";
import UpcomingEvent from "./UpcomingCompetition/UpcomingEvent";
import SpeedcubingFullAnimatedSVG from "./svgs/SpeedcubingFullAnimatedSVG";
import Button from "./ui/Button";
import { SITE_DESCRIPTION } from "../lib/site";

const Hero = () => {
  return (
    <section className="relative flex min-h-[calc(100vh-4rem)] flex-col items-center justify-center overflow-hidden bg-white px-5 py-20 dark:bg-canvas sm:px-8 lg:px-12">
      <div className="pointer-events-none absolute inset-x-0 top-0 h-[9vh] xl:h-[13vh]">
        <div
          className="absolute left-0 top-0 h-full w-[60%] bg-brand-green lg:w-[40%]"
          style={{ clipPath: "polygon(0% 100%, 100% 0%, 0% 0%)" }}
        />
        <div
          className="absolute right-0 top-0 h-full w-[70%] bg-brand-red"
          style={{ clipPath: "polygon(100% 100%, 100% 0%, 0% 0%)" }}
        />
      </div>

      <div className="pointer-events-none absolute inset-x-0 bottom-0 h-[9vh] xl:h-[13vh]">
        <div
          className="absolute bottom-0 left-0 h-full w-[80%] bg-brand-red"
          style={{ clipPath: "polygon(0% 0%, 0% 100%, 100% 100%)" }}
        />
        <div
          className="absolute bottom-0 right-0 h-full w-[60%] bg-brand-green lg:w-[40%]"
          style={{ clipPath: "polygon(100% 0%, 100% 100%, 0% 100%)" }}
        />
      </div>

      <div className="relative z-10 mx-auto flex w-full max-w-6xl flex-col items-center gap-14 xl:flex-row xl:items-center xl:justify-between xl:gap-16">
        <div className="flex max-w-xl flex-col items-center text-center xl:items-start xl:text-left">
          <h1 className="mb-7">
            <span className="sr-only">
              Speedcubing Hungary — a magyar kockázó közösség
            </span>
            <SpeedcubingFullAnimatedSVG animated={true} maxWidth={520} />
          </h1>

          <p className="text-base leading-relaxed text-muted-foreground sm:text-lg">
            {SITE_DESCRIPTION}
          </p>

          <div className="mt-9 flex flex-col gap-3 sm:flex-row">
            <Button
              href="/calendar"
              size="lg"
              icon={<FaCalendar className="h-4 w-4" />}
              iconRight={<FaArrowRight className="h-3.5 w-3.5" />}
            >
              Közelgő versenyek
            </Button>
            <Button
              href="/posts"
              size="lg"
              variant="secondary"
              icon={<FaRegNewspaper className="h-4 w-4" />}
            >
              Friss hírek
            </Button>
          </div>
        </div>

        <UpcomingEvent />
      </div>
    </section>
  );
};

export default Hero;
