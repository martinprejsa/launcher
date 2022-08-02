package manager

import (
	"os"
	"regexp"
	"strconv"
	"strings"
)

func findJAR(dir string) string {
	dfs, _ := os.ReadDir(dir)
	for _, item := range dfs {
		if strings.HasSuffix(item.Name(), ".jar") {
			return item.Name()
		}
	}

	return ""
}

type ProgressBar struct {
	max     int
	current float64
	piece   float64
}

func (p *ProgressBar) SetTaskAmount(amount int) {
	p.max = amount
	p.piece = float64(100) / float64(amount)
}

func (p *ProgressBar) TaskFinished() {
	if p.current+p.piece > float64(p.max) {
		p.current = float64(p.max)
	} else {
		p.current += p.piece
	}
}

type Precision uint8

var (
	CheckMajor    Precision = 1
	CheckMinor    Precision = 2
	CheckPatch    Precision = 4
	PrecisionFull Precision = CheckMajor + CheckMinor + CheckPatch
)

func CompareVersion(required string, current string, precision Precision) bool {
	regex, _ := regexp.Compile("^\\d*.\\d*.\\d*$")
	if regex.MatchString(required) && regex.MatchString(current) {
		var toParts = func(s string) (int, int, int) {
			parts := strings.Split(s, ".")
			major, _ := strconv.Atoi(parts[0])
			minor, _ := strconv.Atoi(parts[1])
			patch, _ := strconv.Atoi(parts[2])
			return major, minor, patch
		}

		var isSet = func(num Precision, mask Precision) bool {
			return (num & mask) != 0
		}

		majorR, minorR, patchR := toParts(required)
		majorC, minorC, patchC := toParts(current)

		ret := true
		if isSet(precision, CheckMajor) {
			if majorR > majorC {
				ret = false
			} else {
				return true
			}
		} else if isSet(precision, CheckMinor) {
			if minorR > minorC {
				ret = false
			} else {
				return true
			}
		} else if isSet(precision, CheckPatch) {
			if patchR > patchC {
				ret = false
			} else {
				return true
			}
		}

		return ret
	} else {
		return false
	}
}
