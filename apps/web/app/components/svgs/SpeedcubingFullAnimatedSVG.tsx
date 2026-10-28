import React from "react";
import {
  LOGO_BOTTOM_COLOR,
  LOGO_BOTTOM_PATH,
  LOGO_CUBE_PATH,
  LOGO_FRAME_PATHS,
  LOGO_LETTER_PATHS,
  LOGO_TOP_COLOR,
  LOGO_TOP_PATH,
} from "./speedcubingLogo";

const SpeedcubingFullAnimatedSVG = ({
  topColor = LOGO_TOP_COLOR,
  bottomColor = LOGO_BOTTOM_COLOR,
  maxWidth = 680,
  animated = true,
}) => {
  return (
    <>
      <svg
        version="1.0"
        className="w-full text-foreground"
        viewBox="80 0 1375.000000 300.000000"
        preserveAspectRatio="xMidYMid meet"
        style={{
          display: "block",
          maxWidth: `${maxWidth}px`,
          height: "auto",
        }}
        xmlns="http://www.w3.org/2000/svg"
      >
        <defs />

        <g
          transform="translate(0.000000,430.000000) scale(0.100000,-0.100000)"
          stroke="none"
        >
          {LOGO_LETTER_PATHS.map((d) => (
            <path key={d} d={d} fill="currentColor" />
          ))}
        </g>

        {!animated && (
          <g
            transform="translate(0.000000,430.000000) scale(0.100000,-0.100000)"
            fill="currentColor"
            stroke="none"
          >
            {LOGO_FRAME_PATHS.map((d) => (
              <path key={d} d={d} />
            ))}
          </g>
        )}

        {!animated && (
          <g>
            <g
              transform="translate(0,430) scale(0.1,-0.1)"
              fill="currentColor"
              stroke="none"
            >
              <path d={LOGO_CUBE_PATH} />
            </g>

            <g
              transform="translate(0.000000,430.000000) scale(0.100000,-0.100000)"
              stroke="none"
            >
              <path
                d={LOGO_TOP_PATH}
                fill={topColor}
              />
              <path
                d={LOGO_BOTTOM_PATH}
                fill={bottomColor}
              />
            </g>
          </g>
        )}

        {animated && (
          <>
            <g
              transform="translate(0.000000,430.000000) scale(0.100000,-0.100000)"
              fill="currentColor"
              stroke="none"
            >
              <path
                className="offset-top opacity-0"
                d="M1640 4050 c-40 -53 -72 -105 -72 -115 0 -17 19 -18 343 -18 189 0 460 -2 603 -5 175 -3 248 -2 224 5 -45 12 -1009 230 -1018 230 -4 0 -40 -44 -80 -97z"
              />

              <path
                className="offset-right opacity-0"
                d="M3500 3600 c-4 -11 -12 -1219 -9 -1215 1 1 11 49 24 107 25 119 198 914 206 945 4 17 -13 35 -95 98 -101 78 -118 87 -126 65z"
              />
              <path
                className="offset-left opacity-0"
                d="M1240 3195 c0 -7 -118 -554 -181 -836 -27 -123 -48 -227 -47 -232 7 -18 201 -157 213 -152 12 4 15 103 17 611 2 334 2 609 1 611 -2 2 -3 2 -3 -2z"
              />
              <path
                className="offset-bottom opacity-0"
                d="M2000 1700 c324 -75 982 -222 984 -219 56 70 146 200 146 212 0 16 -30 17 -342 17 -189 0 -462 2 -608 5 l-265 5 85 -20z"
              />
            </g>

            <g className="rotating-group">
              <g
                transform="translate(0,430) scale(0.1,-0.1)"
                fill="currentColor"
                stroke="none"
              >
                <path d={LOGO_CUBE_PATH} />
              </g>

              <g
                transform="translate(0.000000,430.000000) scale(0.100000,-0.100000)"
                stroke="none"
              >
                <path
                  d={LOGO_TOP_PATH}
                  fill={topColor}
                />
                <path
                  d={LOGO_BOTTOM_PATH}
                  fill={bottomColor}
                />
              </g>
            </g>
          </>
        )}
      </svg>
    </>
  );
};

export default SpeedcubingFullAnimatedSVG;
