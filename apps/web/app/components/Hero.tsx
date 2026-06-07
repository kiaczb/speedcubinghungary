import Image from "next/image";
import Link from "next/link";
import Card from "./Cards/Card";
import { FaCalendar, FaRegNewspaper, FaUser } from "react-icons/fa";
import UpcomingEvent from "./UpcomingCompetition/UpcomingEvent";
import SpeedcubingFullAnimatedSVG from "./svgs/SpeedcubingFullAnimatedSVG";
import SpeedcubingFullSVGugi from "./svgs/SpeedcubingFullSVGugi";

const Hero = () => {
  return (
    <section className="relative flex flex-col bg-white min-h-[93.5vh] xl:flex-row justify-around items-center px-5 py-10 lg:px-40 lg:pb-25 overflow-hidden gap-0 xl:gap-10">
      {/* Top Triangles using clip-path */}
      <div className="absolute top-0 left-0 w-full h-[10vh] xl:h-[15vh]">
        {/* Green triangle - top left, smaller */}
        <div
          className="absolute top-0 left-0 w-[60%] lg:w-[40%] h-full bg-green-600"
          style={{
            clipPath: "polygon(0% 100%, 100% 0%, 0% 0%)",
          }}
        ></div>

        {/* Red triangle - top right, larger */}
        <div
          className="absolute top-0 right-0 w-[70%] h-full bg-red-600"
          style={{
            clipPath: "polygon(100% 100%, 100% 0%, 0% 0%)",
          }}
        ></div>
      </div>

      {/* Bottom Triangles using clip-path */}
      <div className="absolute bottom-0 left-0 w-full h-[10vh] xl:h-[15vh]">
        {/* Green triangle - bottom left, larger */}
        <div
          className="absolute bottom-0 left-0 w-[80%] h-full bg-red-600"
          style={{
            clipPath: "polygon(0% 0%, 0% 100%, 100% 100%)",
          }}
        ></div>

        {/* Red triangle - bottom right, smaller */}
        <div
          className="absolute bottom-0 right-0 w-[60%] lg:w-[40%] h-full bg-green-600"
          style={{
            clipPath: "polygon(100% 0%, 100% 100%, 0% 100%)",
          }}
        ></div>
      </div>

      {/* Rest of your content remains the same */}
      <div className="relative z-10 flex flex-col text-black w-auto p-10 xl:p-0 lg:mb-0">
        <div className="mb-5">
          <SpeedcubingFullAnimatedSVG
            animated={true}
            textColor="black"
            maxWidth={550}
          />
        </div>
        <div>
          <p className="text-base md:text-lg lg:text-xl mb-8 opacity-90 max-w-xl">
            A Speedcubing Hungary Egyesület célja a kockázás népszerűsítése és
            WCA regionális szervezetként hivatalos WCA versenyek
            szervezése.&nbsp;
          </p>

          <div className="flex flex-col sm:flex-row gap-4 mb-8 justify-center xl:justify-start">
            <Link
              href="#events"
              className="btn-primary bg-black text-white font-semibold py-3 px-6 md:py-4 md:px-5 rounded-xl text-base md:text-lg transition duration-300 text-center hover:bg-gray-800"
            >
              <FaCalendar className="inline mr-2 mb-1" />
              Jövőbeli események
            </Link>
            <Link
              href="#join"
              className="border-2 font-semibold py-3 px-6 md:py-4 md:px-5 rounded-xl text-base md:text-lg transition duration-300 text-center hover:bg-gray-200"
            >
              <FaRegNewspaper className="inline mr-2 mb-0.5" />
              Friss hírek
            </Link>
          </div>
        </div>
      </div>
      <UpcomingEvent />
    </section>
  );
};

export default Hero;
