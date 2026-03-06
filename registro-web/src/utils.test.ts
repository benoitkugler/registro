import { expect, test } from "bun:test";
import { FormRules } from "./utils";

test("mail list validator", () => {
  const func = FormRules.validMails();
  expect(func(["ok@free.fr"])).toBe(true);
  expect(func(["notok@free.fr notok@free.fr"])).not.toBe(true);
  expect(func(["l.xxxxx@bluewin.ch so.xxxxx@bluewin.ch"])).not.toBe(true);
});
