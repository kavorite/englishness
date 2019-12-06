package englishness

import (
    "math"
    "strings"
)

func init() {
    MonogramFrequencies.Normalize()
    BigramFrequencies.Normalize()
    TrigramFrequencies.Normalize()
}

func NGrams(doc string, n int, forEach func(string)) {
    doc = strings.ToLower(doc)
    k := len(doc) - n + 1
    for i := 0; i < k; i++ {
        forEach(doc[i:i+n])
    }
}

type NgramFrequencies map[string]float64

func (F NgramFrequencies) Add(t string, delta float64) {
    if _, ok := F[t]; ok {
        F[t] += delta
    } else {
        F[t] = delta
    }
}

func (F NgramFrequencies) Has(t string) bool {
    _, ok := F[t]
    return ok
}

func (F NgramFrequencies) Normalize() {
    var sigma float64
    for k := range F {
        sigma += F[k]
    }
    for k := range F {
        F[k] /= sigma
    }
}

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

func (F NgramFrequencies) MSE(K NgramFrequencies) float64 {
    var sigma float64
    R := F.Residuals(K)
    for k := range R {
        sigma += R[k]*R[k]
    }
    sigma /= float64(len(K))
    return sigma
}

func Eval(doc string) (compoundError float64) {
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
    compoundError = math.Sqrt(mmse*mmse + bmse*bmse + tfse*tfse)
    return
}

func IsEnglish(englishness float64, leniency float64) bool {
    const μ = 0.0004
    const σ = 0.003
    if leniency < 0 {
        leniency = 0.75
    }
    deviance := englishness-μ
    return leniency * math.Abs(deviance) < σ
}
