package englishness

import (
    "math"
    "strings"
    "unicode"

    "golang.org/x/text/transform"
    "golang.org/x/text/unicode/norm"
)

func init() {
    MonogramFrequencies.Normalize()
    BigramFrequencies.Normalize()
    TrigramFrequencies.Normalize()
}

func strip(src string) string {
    strip := func(r rune) bool {
        return unicode.Is(unicode.Mn, r)
    }
    t := transform.Chain(norm.NFD, transform.RemoveFunc(strip), norm.NFC)
    stripped, _, _ := transform.String(t, src)
    return stripped
}

// NGrams runs forEach over the character tuples present in an input text.
func NGrams(doc string, n int, forEach func(string)) {
    doc = strings.ToLower(doc)
    k := len(doc) - n + 1
    for i := 0; i < k; i++ {
        forEach(doc[i:i+n])
    }
}

// NgramFrequencies maps English character tuples to their number of
// occurrences.
type NgramFrequencies map[string]float64

// Add adds delta to the frequency corresponding to t. If no t frequency is
// present, it is assumed zero.
func (F NgramFrequencies) Add(t string, delta float64) {
    if _, ok := F[t]; ok {
        F[t] += delta
    } else {
        F[t] = delta
    }
}

// Has returns whether t ∈ F.
func (F NgramFrequencies) Has(t string) bool {
    _, ok := F[t]
    return ok
}

// Normalize places all the values in F on the closed interval [0, 1].
func (F NgramFrequencies) Normalize() {
    var sigma float64
    for k := range F {
        sigma += F[k]
    }
    for k := range F {
        F[k] /= sigma
    }
}

// Residuals returns the difference of the values of the common keys found in F
// and K.
func (F NgramFrequencies) Residuals(K NgramFrequencies) (R NgramFrequencies) {
    n := len(F)
    if len(K) < n {
        n = len(K)
    }
    R = make(NgramFrequencies, n)
    for t := range F {
        if K.Has(t) {
            R.Add(t, F[t])
            R.Add(t, -K[t])
        }
    }
    return
}

// MSE returns the mean squared error of the residuals between F and K.
func (F NgramFrequencies) MSE(K NgramFrequencies) float64 {
    var sigma float64
    R := F.Residuals(K)
    for k := range R {
        sigma += R[k]*R[k]
    }
    sigma /= float64(len(K))
    return sigma
}

func Eval(doc string) (englishness float64) {
    mf := make(NgramFrequencies, len(MonogramFrequencies))
    bf := make(NgramFrequencies, len(BigramFrequencies))
    tf := make(NgramFrequencies, len(TrigramFrequencies))
    NGrams(doc, 3, func(trigram string) {
        tf.Add(trigram, 1)
        NGrams(trigram, 2, func(bigram string) {
            bf.Add(bigram, 1)
            NGrams(bigram, 1, func(monogram string) {
                mf.Add(monogram, 1)
            })
        })
    })
    mf.Normalize()
    bf.Normalize()
    tf.Normalize()
    mmse := mf.MSE(MonogramFrequencies)
    bmse := bf.MSE(BigramFrequencies)
    tfse := tf.MSE(TrigramFrequencies)
    englishness = math.Sqrt(mmse*mmse + bmse*bmse + tfse*tfse)
    return
}

// IsEnglish returns whether the given englishness value is within 2/strictness
// standard deviations of the median englishness found by an empirical
// determination. If strictness is negative, a reasonable default is used.
func IsEnglish(englishness float64, strictness float64) bool {
    const μ = 0.00035
    const σ = 0.002
    if strictness < 0 {
        strictness = 0.85
    }
    deviance := englishness-μ
    return math.Abs(deviance) < 2*σ / strictness
}
