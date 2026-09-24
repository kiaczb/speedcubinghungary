package middleware

import "net/http"

// SecurityHeaders az API minden válaszára ráteszi a tartalomtípus-találgatás
// tiltását.
//
// ⚠️ Ez a `/uploads/*` miatt számít igazán: ott FELTÖLTÖTT bájtokat szolgálunk
// ki. A típust ugyan a fájl tartalmából állapítjuk meg (§17), de egy
// kép-és-HTML kettős tartalmú fájlt a sniffelő böngésző dokumentumként is
// értelmezhetne — méghozzá az API saját origin-jén. A nosniff ezt zárja ki.
//
// A frontend fejléceit (CSP, frame-ancestors, Referrer-Policy) a Next adja,
// lásd apps/web/next.config.ts — azok a HTML válaszokra valók, ez viszont a
// Next elé nem kerülő API-válaszokra.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, r)
	})
}
