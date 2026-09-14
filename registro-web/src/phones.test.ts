import { expect, test } from "bun:test";
import { _lookupIndicatif, Phones } from "./phones";

test("lookupIndicatif", () => {
  expect(_lookupIndicatif("")).toBeUndefined();
  expect(_lookupIndicatif("0685194578")).toBeUndefined();
  expect(_lookupIndicatif("+")).toBeUndefined();
  expect(_lookupIndicatif("+2")).toBeUndefined();
  expect(_lookupIndicatif("+1")).toEqual([2, "US"]);
  expect(_lookupIndicatif("001")).toEqual([3, "US"]);
  expect(_lookupIndicatif("+123789")).toEqual([2, "US"]);
  expect(_lookupIndicatif("+33456")).toEqual([3, "FR"]);
  expect(_lookupIndicatif("+41")).toEqual([3, "CH"]);
  expect(_lookupIndicatif("+413")).toEqual([3, "CH"]);
  expect(_lookupIndicatif("+41684958956")).toEqual([3, "CH"]);
  expect(_lookupIndicatif("00330652")).toEqual([4, "FR"]);
  expect(_lookupIndicatif("0041684958956")).toEqual([4, "CH"]);
});

test("Phones._parse", () => {
  expect(Phones.parse("")).toEqual({
    indicatif: "",
    localNumber: "",
    flag: "",
    tel: "",
    country: null,
  });
  expect(Phones.parse("0685194578")).toEqual({
    indicatif: "",
    localNumber: "0685194578",
    flag: "",
    tel: "0685194578",
    country: null,
  });
  expect(Phones.parse("+")).toEqual({
    indicatif: "",
    localNumber: "+",
    flag: "",
    tel: "+",
    country: null,
  });
  expect(Phones.parse("+2")).toEqual({
    indicatif: "",
    localNumber: "+2",
    flag: "",
    tel: "+2",
    country: null,
  });
  expect(Phones.parse("+1")).toEqual({
    indicatif: "+1",
    localNumber: "",
    flag: "🇺🇸",
    tel: "+1",
    country: "US",
  });
  expect(Phones.parse("+123789")).toEqual({
    indicatif: "+1",
    localNumber: "23789",
    flag: "🇺🇸",
    tel: "+123789",
    country: "US",
  });
  expect(Phones.parse("+33456")).toEqual({
    indicatif: "+33",
    localNumber: "456",
    flag: "🇫🇷",
    tel: "+33456",
    country: "FR",
  });
  expect(Phones.parse("+41")).toEqual({
    indicatif: "+41",
    localNumber: "",
    flag: "🇨🇭",
    tel: "+41",
    country: "CH",
  });
  expect(Phones.parse("+413")).toEqual({
    indicatif: "+41",
    localNumber: "3",
    flag: "🇨🇭",
    tel: "+413",
    country: "CH",
  });
  expect(Phones.parse("+41684958956")).toEqual({
    indicatif: "+41",
    localNumber: "684958956",
    flag: "🇨🇭",
    tel: "+41684958956",
    country: "CH",
  });
});

test("format", () => {
  expect(Phones.format("0655125678")).toBe("06 55 12 56 78");
  expect(Phones.format("+330655125678")).toBe("+33 06 55 12 56 78");
  expect(Phones.format("+33 0 76 61 96 06")).toBe("+33 07 66 19 60 6");
  expect(Phones.format("+33655125678")).toBe("+33 6 55 12 56 78");
  expect(Phones.format("+41065512567823")).toBe("+41 065 512 56 7823");
  expect(Phones.format("+410655125678")).toBe("+41 065 512 56 78");
  expect(Phones.format("+41655125678")).toBe("+41 65 512 56 78");
  expect(Phones.format("+41655")).toBe("+41 65 5");
  expect(Phones.format("+416554")).toBe("+41 65 54");
  expect(Phones.format("+4106554")).toBe("+41 065 54");
  expect(Phones.format("00330655125678")).toBe("0033 06 55 12 56 78");
  expect(Phones.format("004106554")).toBe("0041 065 54");
});
