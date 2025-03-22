import { expect, test } from "vitest";
import { ageFrom, autocomplete, isDateZero, parse } from "./date";
import type { Date_ } from "@/clients/inscription/logic/api";

test("parse", () => {
  expect(parse("")).toBe(undefined);
  expect(parse("1/2/20")).toBe(undefined);
  expect(parse("1/002/20")).toBe(undefined);
  expect(parse("1/2/2000")).toBe("2000-02-01");
  expect(parse("01/03/2000")).toBe("2000-03-01");
  expect(parse("1/12/2000")).toBe("2000-12-01");
  expect(parse("1-12-2000")).toBe("2000-12-01");
  expect(parse("01-03-2000")).toBe("2000-03-01");
});

test("autocomplete", () => {
  expect(autocomplete("")).toBe("");
  expect(autocomplete("1")).toBe("1");
  expect(autocomplete("12")).toBe("12/");
  expect(autocomplete("12/01")).toBe("12/01/");
  expect(autocomplete("1/01")).toBe("1/01/");
  expect(autocomplete("/2004")).toBe("/2004");
});

test("isDateZero", () => {
  expect(isDateZero("" as Date_)).toBe(true);
  expect(isDateZero("54" as Date_)).toBe(true);
  expect(isDateZero("0001-01-01" as Date_)).toBe(true);
  expect(isDateZero("2000-01-01" as Date_)).toBe(false);
  expect(isDateZero("1900-01-01" as Date_)).toBe(false);
});

test("age", () => {
  expect(ageFrom("" as Date_)).toBe(null);
  expect(ageFrom("" as Date_, new Date(Date.now()))).toBe(null);
  expect(ageFrom("2000-01-01" as Date_, new Date("2000-01-02"))).toBe(0);
  expect(ageFrom("2000-01-01" as Date_, new Date("2002-01-01"))).toBe(2);
  expect(ageFrom("2000-05-01" as Date_, new Date("2002-04-02"))).toBe(1);
  expect(ageFrom("2000-02-29" as Date_, new Date("2002-02-28"))).toBe(1);
  expect(ageFrom("2000-02-29" as Date_, new Date("2002-03-01"))).toBe(2);
  expect(ageFrom("2001-02-28" as Date_, new Date("2004-02-29"))).toBe(3);
});
