package main

import (
    "bytes"
    "strings"
    //"mygi"
    //"fmt"
)

func strSliceContains(slice []string,str string) bool {
    for _, v := range slice {
        if str == v {
            return true
        }
    }
    return false
}


// snake_case to CamelCase
func snake2Camel(name string) string {
    //name = strings.ToLower(name)
	var out bytes.Buffer
	for _, word := range strings.Split(name, "_") {
		word = strings.ToLower(word)
		//if subst, ok := config.word_subst[word]; ok {
			//out.WriteString(subst)
			//continue
		//}

		if word == "" {
			out.WriteString("_")
			continue
		}
		out.WriteString(strings.ToUpper(word[0:1]))
		out.WriteString(word[1:])
	}
	return out.String()
}


