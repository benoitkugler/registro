package ancv

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"strings"
)

// seal computes the ANCV signature of an ordered list of fields.
//
// Kit §5.3.2.1 (p. 53) : « les valeurs des données de cette liste et les
// concatène dans une chaîne de caractère avec le séparateur &. [...] Si une
// valeur de ces listes est vide, alors elle n'est pas utilisée dans la
// concaténation : le séparateur & ne peut être présent ni en début de chaîne,
// ni en fin de chaîne, ni doublé dans la chaîne. »
// Kit §5.3.2.2 (p. 53) : « signe la chaîne concaténée [...] avec l'algorithme
// HmacSHA256 [...] encodée en base64 SafeUrl » et les chaînes « doivent être
// en format UTF-8 ».
func seal(key string, fields ...string) string {
	values := make([]string, 0, len(fields))
	for _, field := range fields {
		if field != "" {
			values = append(values, field)
		}
	}

	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(strings.Join(values, "&")))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// securityHeader formats the ANCV-Security header.
// Kit §5.3.2.3 (p. 53) : « algorithme.versionKey.chaineScellée », par exemple
// « HmacSHA256.version-3620.f536975d06c0309214f805bb90ccff089ef ».
// Kit §4.3.1 (p. 13) donne le nom du header : « Ancv-Security ».
func securityHeader(version, seal string) string {
	return "HmacSHA256." + version + "." + seal
}

// initSealFields returns the sealed fields of POST /payment-transactions.
// Kit §4.5.1.4 (p. 18) : 1 Merchant → shopId, 2 Merchant → serviceProviderId
// (uniquement si le paramètre est transmis), 3 Order → id, 4 Order → paymentId,
// 5 Order → Amount → total. serviceProviderId n'est pas transmis (intégration
// directe), donc le champ est vide et ignoré par [seal].
func initSealFields(merchant merchant, order Order) []string {
	return []string{
		formatShopID(merchant.ShopID),
		"", // serviceProviderId : intégration directe
		order.ID,
		order.PaymentID,
		formatAmount(order.Amount.Total),
	}
}

// paySealFields returns the sealed fields of POST
// /payment-transactions/{id}/payer.
// Kit §4.5.2.4 (p. 21) : 1 Transaction → id, 2 Payer → beneficiaryId,
// 3 Payer → Amount → total.
func paySealFields(transactionID string, payer payer) []string {
	return []string{
		transactionID,
		payer.BeneficiaryID,
		formatAmount(payer.Amount.Total),
	}
}

// statusSealFields returns the sealed fields of GET /payment-transactions/{id}.
// Kit §4.5.3.3 (p. 24) : « 1 transaction → id ».
func statusSealFields(transactionID string) []string { return []string{transactionID} }
