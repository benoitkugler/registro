const reSepTel = /[ -/;\t]/g;

function splitBySize(a: string) {
  const b = [];
  for (var i = 2; i < a.length; i += 2) {
    // length 2, for example
    b.push(a.slice(i - 2, i));
  }
  b.push(a.slice(a.length - (2 - (a.length % 2)))); // last fragment
  return b;
}

export function formatTel(tel: string) {
  tel = tel.replace(reSepTel, "");
  if (tel.length < 8) {
    return splitBySize(tel).join(" ");
  } // numéro incomplet, on insert des espaces
  const start = tel.length - 8; // 8 derniers chiffres
  const chunks = [tel.substr(0, start)];
  for (let i = 0; i < 4; i++) {
    chunks.push(tel.substr(start + 2 * i, 2));
  }
  return chunks.join(" ");
}

export function formatDate(date: Date, showYear = false, showWeekday = true) {
  if (isNaN(date.valueOf()) || date.getFullYear() <= 1) {
    return "";
  }
  const s = date.toLocaleString(undefined, {
    year: showYear ? "numeric" : undefined,
    day: "numeric",
    month: "short",
    // hour: "2-digit",
    // minute: showMinute ? "2-digit" : undefined,
  });
  if (showWeekday) {
    return `${_weekdays[date.getDay()]} ${s}`;
  }
  return s;
}

const _weekdays = ["Dim.", "Lun.", "Mar.", "Mer.", "Jeu.", "Ven.", "Sam."];
