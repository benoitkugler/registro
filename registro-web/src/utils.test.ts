import { expect, test } from "bun:test";
import { Camps, FormRules } from "./utils";

test("quotient familial grid", () => {
  const defaultGrid = Camps.defaultQuotientFamilial();
  expect(Camps.isQuotientFamilialActive([])).toBe(false);
  expect(Camps.isQuotientFamilialActive(null)).toBe(false);
  expect(Camps.isQuotientFamilialActive(defaultGrid)).toBe(true);
  expect(defaultGrid[defaultGrid.length - 1].Pourcentage as number).toBe(100);
});

test("mail list validator", () => {
  const func = FormRules.validMails();
  expect(func(["ok@free.fr"])).toBe(true);
  expect(func(["notok@free.fr notok@free.fr"])).not.toBe(true);
  expect(func(["l.xxxxx@bluewin.ch so.xxxxx@bluewin.ch"])).not.toBe(true);
});
