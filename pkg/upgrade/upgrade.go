package upgrade

import (
	"fmt"
	"unicode"

	"github.com/Jguer/yay/v11/pkg/db"
	"github.com/Jguer/yay/v11/pkg/intrange"
	"github.com/Jguer/yay/v11/pkg/text"
)

// Filter decides if specific package should be included in theincluded in the  results.
type Filter func(Upgrade) bool

// Upgrade type describes a system upgrade.
type Upgrade = db.Upgrade

func StylizedNameWithRepository(u Upgrade) string {
	return text.Bold(text.ColorHash(u.Repository)) + "/" + text.Bold(u.Name)
}

// upSlice is a slice of Upgrades.
type UpSlice struct {
	Up    []Upgrade
	Repos []string
}

func (u UpSlice) Len() int      { return len(u.Up) }
func (u UpSlice) Swap(i, j int) { u.Up[i], u.Up[j] = u.Up[j], u.Up[i] }

func (u UpSlice) Less(i, j int) bool {
	if u.Up[i].Repository == u.Up[j].Repository {
		iRunes := []rune(u.Up[i].Name)
		jRunes := []rune(u.Up[j].Name)

		return text.LessRunes(iRunes, jRunes)
	}

	for _, db := range u.Repos {
		if db == u.Up[i].Repository {
			return true
		} else if db == u.Up[j].Repository {
			return false
		}
	}

	iRunes := []rune(u.Up[i].Repository)
	jRunes := []rune(u.Up[j].Repository)

	return text.LessRunes(iRunes, jRunes)
}

func GetVersionDiff(oldVersion, newVersion string) (left, right string) {
	if oldVersion == newVersion {
		return oldVersion + text.Red(""), newVersion + text.Green("")
	}

	checkWords := func(str string, index int, words ...string) bool {
		for _, word := range words {
			wordLength := len(word)

			nextIndex := index + 1
			if (index < len(str)-wordLength) &&
				(str[nextIndex:(nextIndex+wordLength)] == word) {
				return true
			}
		}

		return false
	}

	diffPosition := intrange.Min(len(oldVersion), len(newVersion))
	for index, char := range oldVersion {
		if index >= len(newVersion) || char != rune(newVersion[index]) {
			diffPosition = index
			break
		}
	}
	colorize := func(version string, colorFunc func(string) string) (out string) {
		charIsSpecial := func(char byte) bool {
			return !(unicode.IsLetter(rune(char)) || unicode.IsNumber(rune(char)))
		}
		diffComponentStart := 0
		if charIsSpecial(version[diffPosition]) {
			diffComponentStart = diffPosition
		} else {
			for index := diffPosition - 1; index >= 0; index-- {
				if charIsSpecial(version[index]) || checkWords(oldVersion, index, "rc", "pre", "alpha", "beta") {
					diffComponentStart = index + 1
					break
				}
			}
		}
		out = version[:diffComponentStart]
		offset := 0
		for index, char := range version[diffComponentStart:] {
			if index > 0 && charIsSpecial(byte(char)) {
				out += colorFunc(version[diffComponentStart+offset : diffComponentStart+index])
				out += string(char)
				offset = index + 1
			}
		}
		return out + colorFunc(version[diffComponentStart+offset:])
	}
	return colorize(oldVersion, text.Red), colorize(newVersion, text.Green)
}

// Print prints the details of the packages to upgrade.
func (u UpSlice) Print() {
	longestName, longestVersion := 0, 0

	for _, pack := range u.Up {
		packNameLen := len(StylizedNameWithRepository(pack))
		packVersionLen := len(pack.LocalVersion)
		longestName = intrange.Max(packNameLen, longestName)
		longestVersion = intrange.Max(packVersionLen, longestVersion)
	}

	namePadding := fmt.Sprintf("%%-%ds  ", longestName)
	numberPadding := fmt.Sprintf("%%%dd  ", len(fmt.Sprintf("%v", len(u.Up))))

	for k, i := range u.Up {
		left, right := GetVersionDiff(i.LocalVersion, i.RemoteVersion)

		fmt.Print(text.Magenta(fmt.Sprintf(numberPadding, len(u.Up)-k)))

		fmt.Printf(namePadding, StylizedNameWithRepository(i))
		padding := fmt.Sprintf(fmt.Sprintf("%%-%ds", longestVersion-len(i.LocalVersion)), "")
		fmt.Printf("%-s%s -> %s\n", left, padding, right)
	}
}
