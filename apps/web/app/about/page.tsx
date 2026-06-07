import SpeedcubingLogoSVG from "../components/svgs/SpeedcubingLogoSVG";

export default function About() {
  return (
    <div className="p-4">
      <div className="flex justify-center">
        <h1 className="text-4xl font-bold mb-10">Rólunk</h1>
      </div>
      <div className="flex flex-col md:flex-row md:justify-around">
        <div className="w-full p-5">
          <h2 className="font-bold text-2xl mb-3">Speedcubing Hungary</h2>
          <p className="mb-4 max-w-lg">
            Kérdéseid vannak a kockázással kapcsolatban? Szeretnél versenyt
            szervezni vagy részt venni egy eseményen? Vedd fel velünk a
            kapcsolatot e-mailben vagy a honlapunkon!
          </p>
        </div>
      </div>
    </div>
  );
}
