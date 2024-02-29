package main

import (
    "strings"
)

func wordPattern(pattern string, s string) bool {
    pDict := make(map[rune]string)
    idx := 0
    s += " "
    for _, c := range pattern {
        if word, exists := pDict[c]; exists {
            nextSpaceIndex := strings.Index(s[idx:], " ") + idx
            if word != s[idx:nextSpaceIndex] {
                return false
            }
            idx = nextSpaceIndex + 1
        } else {
            nextSpaceIndex := strings.Index(s[idx:], " ") + idx
            for _, v := range pDict {
                if v == s[idx:nextSpaceIndex] {
                    return false
                }
            }
            pDict[c] = s[idx:nextSpaceIndex]
            idx = nextSpaceIndex + 1
        }
    }
    return idx == len(s)
}

func wordPattern2(pattern string, s string) bool {
    pDict := make(map[rune]string)
    words := strings.Split(s, " ")
    pSet := make(map[rune]struct{})
    wSet := make(map[string]struct{})

    for _, c := range pattern {
        pSet[c] = struct{}{}
    }
    for _, word := range words {
        wSet[word] = struct{}{}
    }

    if len(pSet) != len(wSet) || len(pattern) != len(words) {
        return false
    }

    for i, c := range pattern {
        if word, exists := pDict[c]; exists {
            if word != words[i] {
                return false
            }
        } else {
            pDict[c] = words[i]
        }
    }

    return true
}
