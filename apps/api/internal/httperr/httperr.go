// Package httperr egyetlen helyre gyűjti, hogyan lesz egy hibából HTTP válasz.
//
// A szabály egyetlen mondat: a hiba SZÖVEGE csak 4xx-nél megy ki a kliensnek.
//
// 4xx a hívó hibája, és a mi saját sentinel hibáink írják le ("post not found",
// "invalid round number") — a hívónak épp az a dolga, hogy megtudja, mit
// rontott el. 5xx viszont a MI hibánk, és a szövege a belső működésről beszél:
// egy nyers `pq:` hiba a séma tábla- és oszlopneveit adja ki, egy `*url.Error`
// a belső upstream címeket (a WCA Live/WCA API hostját és a query stringben
// utazó azonosítókat), egy `net.OpError` pedig a Docker-hálózat neveit és
// portjait. Ezek a böngészőben semmit nem segítenek, a támadónak viszont
// ingyen felderítés.
//
// Ezért 5xx-nél a részletes hiba a SZERVER naplójába megy, a válasz pedig egy
// általános, státusz szerinti üzenet.
package httperr

import (
	"log"
	"net/http"
)

// Write kiírja a hibát a válaszba — 5xx esetén a naplóba is.
//
// A status megadása a hívóé: ő tudja, melyik sentinel hibája mit jelent
// (lásd pl. post.ErrPostNotFound → 404). Ez a függvény csak azt dönti el,
// hogy a szöveg kimehet-e.
func Write(w http.ResponseWriter, r *http.Request, status int, err error) {
	if status < http.StatusInternalServerError {
		http.Error(w, err.Error(), status)
		return
	}

	log.Printf("[http] %s %s -> %d: %v", r.Method, r.URL.Path, status, err)
	http.Error(w, publicMessage(status), status)
}

// publicMessage az az üzenet, amit 5xx-nél a kliens lát. Szándékosan
// tartalmatlan: annyit mond meg, kinél van a hiba, és hogy érdemes-e
// újrapróbálni.
func publicMessage(status int) string {
	switch status {
	case http.StatusBadGateway:
		return "upstream service error"
	case http.StatusServiceUnavailable:
		return "service temporarily unavailable"
	default:
		return "internal server error"
	}
}
