import type { Date_ } from "@/clients/inscription/logic/api";

const isFull = /^(\d{1,2})\/(\d{1,2})\/(\d{4})$/;
/** parse expects a JJ/MM/AAAA format */
export function parse(s: string) {
  const match = s.match(isFull);
  if (!match) return;

  const day = match[1];
  const month = match[2];
  const year = match[3];
  const date = `${year}-${month.padStart(2, "0")}-${day.padStart(2, "0")}`;
  return date as Date_;
}

const isTwoDigits = /^\d{2}$/;
const isTwoSlashTwoDigits = /^\d{1,2}\/\d{2}$/;

/** autocomplete add a trailing / when appropriate */
export function autocomplete(s: string) {
  if (isTwoDigits.test(s) || isTwoSlashTwoDigits.test(s)) {
    return s + "/";
  }
  return s;
}

export function isDateZero(d: Date_) {
  if (d.length != 10) return true;
  if (d == "0001-01-01") return true;
  return isNaN(new Date(d).valueOf());
}

export function ageFrom(d: Date_, now?: Date) {
  now = now || new Date(Date.now());
  if (isNaN(now.valueOf()) || isDateZero(d)) return null;

  const naissance = new Date(d);
  const diff = now.getFullYear() - naissance.getFullYear();
  const isNaissanceBefore =
    naissance.getMonth() < now.getMonth() ||
    (naissance.getMonth() == now.getMonth() &&
      naissance.getDate() <= now.getDate());
  if (isNaissanceBefore) {
    return diff;
  }
  return diff - 1;
}

export function newDate_(d: Date) {
  const offset = d.getTimezoneOffset();
  d = new Date(d.getTime() - offset * 60 * 1000);
  return d.toISOString().split("T")[0] as Date_;
}
