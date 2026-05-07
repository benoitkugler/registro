import { expect, test } from "bun:test";
import { lookupIndicatif, parseTel } from "./phones";

test("lookupIndicatif", () => {
  expect(lookupIndicatif("")).toBeUndefined();
  expect(lookupIndicatif("0685194578")).toBeUndefined();
  expect(lookupIndicatif("+")).toBeUndefined();
  expect(lookupIndicatif("+2")).toBeUndefined();
  expect(lookupIndicatif("+1")).toEqual([2, "US"]);
  expect(lookupIndicatif("+123789")).toEqual([2, "US"]);
  expect(lookupIndicatif("+33456")).toEqual([3, "FR"]);
  expect(lookupIndicatif("+41")).toEqual([3, "CH"]);
  expect(lookupIndicatif("+413")).toEqual([3, "CH"]);
  expect(lookupIndicatif("+41684958956")).toEqual([3, "CH"]);
});

test("parseTel", () => {
  expect(parseTel("")).toEqual({ indicatif: "", number: "" });
  expect(parseTel("0685194578")).toEqual({
    indicatif: "",
    number: "0685194578",
  });
  expect(parseTel("+")).toEqual({ indicatif: "", number: "+" });
  expect(parseTel("+2")).toEqual({ indicatif: "", number: "+2" });
  expect(parseTel("+1")).toEqual({ indicatif: "+1", number: "" });
  expect(parseTel("+123789")).toEqual({ indicatif: "+1", number: "23789" });
  expect(parseTel("+33456")).toEqual({ indicatif: "+33", number: "456" });
  expect(parseTel("+41")).toEqual({ indicatif: "+41", number: "" });
  expect(parseTel("+413")).toEqual({ indicatif: "+41", number: "3" });
  expect(parseTel("+41684958956")).toEqual({
    indicatif: "+41",
    number: "684958956",
  });
});
