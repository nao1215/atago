// Package plural renders counted nouns for the messages atago shows a person:
// "1 page" rather than "1 pages", "2 entries" rather than "2 entry".
//
// It exists because the same counted nouns appear in assertions, in the CLI's
// own output, and in generated documentation, and each place had spelled them
// differently — some pluralized, some always plural, some hedged with "(s)".
// A reader should not have to notice which part of atago wrote a sentence.
package plural

import "strconv"

// Count renders n with the noun that fits it: Count(1, "page", "pages") is
// "1 page" and Count(0, "page", "pages") is "0 pages". Only ±1 takes the
// singular, which is what English does with counted nouns.
func Count(n int, singular, plural string) string {
	return strconv.Itoa(n) + " " + Noun(n, singular, plural)
}

// Noun returns the form of the noun that fits n, without the number. Use it
// when the count is already in the sentence for another reason.
func Noun(n int, singular, plural string) string {
	if n == 1 || n == -1 {
		return singular
	}
	return plural
}
