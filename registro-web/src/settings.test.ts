import { expect, test } from "bun:test";
import { Currency } from "./clients/backoffice/logic/api";
import {
  currencyItemsFor,
  parseSupportedCurrencies,
  singleCurrency,
} from "./settings";

test("parse supported currencies", () => {
  expect(parseSupportedCurrencies("0")).toEqual([Currency.Euros]);
  expect(parseSupportedCurrencies("0,1")).toEqual([
    Currency.Euros,
    Currency.FrancsSuisse,
  ]);
  expect(parseSupportedCurrencies(" 1 , 0 ")).toEqual([
    Currency.FrancsSuisse,
    Currency.Euros,
  ]);
  expect(parseSupportedCurrencies(undefined)).toEqual([
    Currency.Euros,
    Currency.FrancsSuisse,
  ]);
  expect(() => parseSupportedCurrencies("")).toThrow();
  expect(() => parseSupportedCurrencies("0,2")).toThrow();
  expect(() => parseSupportedCurrencies("CHF")).toThrow();
});

test("currency items follow the supported currencies", () => {
  expect(currencyItemsFor([Currency.Euros])).toEqual([{ value: 0, title: "€" }]);
  expect(currencyItemsFor([Currency.Euros, Currency.FrancsSuisse])).toEqual([
    { value: 0, title: "€" },
    { value: 1, title: "CHF" },
  ]);
  // No VITE_SUPPORTED_CURRENCIES in the test environment: both are allowed.
  expect(singleCurrency).toBe(false);
});
