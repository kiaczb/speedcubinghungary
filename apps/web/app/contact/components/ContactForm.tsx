import React from "react";
import Button from "@/app/components/ui/Button";
import Field, { controlClass } from "@/app/components/ui/Field";

/**
 * Kapcsolat-űrlap.
 *
 * A négy mező korábban négyszer bemásolt, azonos felépítésű blokk volt, és a
 * címkék egyikéhez sem tartozott `htmlFor` — így kattintásra nem ugrott a
 * mezőre a fókusz, és képernyőolvasóval sem volt egyértelmű, melyik címke
 * melyik mezőhöz tartozik. A `Field` ezt kötelezővé teszi.
 *
 * A magyar felület mellett a feliratok is magyarok lettek (korábban angolul
 * álltak egy egyébként végig magyar oldalon).
 *
 * MEGJEGYZÉS: az űrlapnak továbbra sincs backendje — a küldés nincs bekötve.
 */
const ContactForm = () => {
  return (
    <form className="flex flex-col gap-5">
      <Field label="Neved" htmlFor="contact-name">
        <input
          type="text"
          id="contact-name"
          name="name"
          autoComplete="name"
          className={controlClass}
          placeholder="Kiss Anna"
        />
      </Field>

      <Field label="E-mail-címed" htmlFor="contact-email">
        <input
          type="email"
          id="contact-email"
          name="email"
          autoComplete="email"
          className={controlClass}
          placeholder="anna@example.com"
        />
      </Field>

      <Field label="Tárgy" htmlFor="contact-subject">
        <input
          type="text"
          id="contact-subject"
          name="subject"
          className={controlClass}
          placeholder="Miben segíthetünk?"
        />
      </Field>

      <Field label="Üzeneted" htmlFor="contact-message">
        <textarea
          id="contact-message"
          name="message"
          rows={6}
          className={`${controlClass} resize-y`}
          placeholder="Írd le, mivel kereshetünk meg…"
        />
      </Field>

      <Button type="submit" size="lg" fullWidth>
        Üzenet küldése
      </Button>
    </form>
  );
};

export default ContactForm;
