import { Currency, CurrencyLabels } from "./clients/backoffice/logic/api";
import { selectItems } from "./utils";

const knownCurrencies: Currency[] = Object.values(Currency);

/**
 * Parses the comma-separated currency codes from VITE_SUPPORTED_CURRENCIES.
 * Kept as a function so it can be tested independently of the current build mode.
 */
export function parseSupportedCurrencies(
  value: string | undefined,
): Currency[] {
  if (value === undefined) {
    // The variable is always set by the per-asso .env files. Fall back to
    // allowing both currencies when it is missing (e.g. in unit tests).
    return [...knownCurrencies];
  }

  const out = new Set<Currency>();
  for (const part of value.split(",")) {
    const raw = part.trim();
    const currency = Number(raw);
    if (raw === "" || !knownCurrencies.some((known) => known === currency)) {
      throw new Error(`Invalid supported currency: '${raw}'`);
    }
    out.add(currency as Currency);
  }
  return [...out];
}

export function currencyItemsFor(currencies: Currency[]) {
  return selectItems(CurrencyLabels).filter((item) =>
    currencies.some((allowed) => allowed === item.value),
  );
}

const supportedCurrencies = parseSupportedCurrencies(
  import.meta.env.VITE_SUPPORTED_CURRENCIES,
);

export const currencyItems = currencyItemsFor(supportedCurrencies);

/** When only one currency is available, MontantField can hide the selector. */
export const singleCurrency = currencyItems.length === 1;
