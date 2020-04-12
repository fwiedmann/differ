/*
 * MIT License
 *
 * Copyright (c) 2019 Felix Wiedmann
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package tags_analyzer

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"

	log "github.com/sirupsen/logrus"
)

var onlyDigits = "\\d+"
var numberRegex = regexp.MustCompile(onlyDigits)

// GetRegexExprForTag creates a regex expression dynamically for the given tag. It try's to exchange the digits with the onyDigits expression.
func GetRegexExprForTag(tag string) (*regexp.Regexp, error) {
	repl := numberRegex.ReplaceAllString(tag, onlyDigits)
	return regexp.Compile(fmt.Sprintf("^%s$", repl))
}

// GetLatestTagWithRegexExpr filters valid tags for the given expression and sort those. The latest valid tag will be returned
func GetLatestTagWithRegexExpr(tags []string, regx *regexp.Regexp) (string, error) {
	tagsToSort := make(sorter, 0)
	for _, tag := range tags {
		if regx.MatchString(tag) {
			tagDigits, err := getDigitsFromString(tag)
			if err != nil {
				log.Warn(err)
				continue
			}
			tagsToSort = append(tagsToSort, tagInfo{
				digits:   tagDigits,
				complete: tag,
			})
		}
	}
	sort.Sort(tagsToSort)
	if len(tagsToSort) == 0 {
		return "", fmt.Errorf("tags-analyzer: could not find any valid tags with pattern %s from tags %s", regx.String(), tags)
	}
	return tagsToSort[len(tagsToSort)-1].complete, nil
}

func getDigitsFromString(str string) ([]int, error) {
	var convertedNumbers []int
	found := numberRegex.FindAllString(str, -1)
	for _, strNumber := range found {
		number, err := strconv.Atoi(strNumber)
		if err != nil {
			return nil, fmt.Errorf("tags-analyzer: could not get version numbers from tagInfo %s", str)
		}
		convertedNumbers = append(convertedNumbers, number)
	}
	return convertedNumbers, nil
}
