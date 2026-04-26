package handlers

import (
	"encoding/binary"
	"sort"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

func parseModulesFromMOFBytes(mofBytes []byte) []map[string]interface{} {
	modules := []map[string]interface{}{}
	if len(mofBytes) == 0 {
		return modules
	}

	mofStr := decodeMOFText(mofBytes)
	if mofStr == "" {
		return modules
	}

	modMap := map[string]map[string]interface{}{}
	currentModuleName := ""

	for _, line := range strings.Split(mofStr, "\n") {
		line = strings.TrimSpace(strings.TrimSuffix(line, "\r"))
		line = strings.ReplaceAll(line, ";", "")

		switch {
		case strings.HasPrefix(line, "ModuleName = "):
			name := strings.Trim(line[len("ModuleName = "):], " \"'{}[];")
			currentModuleName = name
			if name != "" {
				if _, ok := modMap[name]; !ok {
					modMap[name] = map[string]interface{}{"name": name}
				}
			}

		case strings.HasPrefix(line, "ModuleVersion = "):
			version := strings.Trim(line[len("ModuleVersion = "):], " \"'{}[];")
			if currentModuleName != "" && version != "" {
				modMap[currentModuleName]["version"] = version
			}

		case strings.HasPrefix(line, "RequiredModules = "):
			depsRaw := strings.Trim(line[len("RequiredModules = "):], " {}\"'[];")
			if currentModuleName == "" || depsRaw == "" {
				continue
			}
			depList := []string{}
			for _, dep := range strings.Split(depsRaw, ",") {
				dep = strings.Trim(dep, " \"'{}[];")
				if dep != "" {
					depList = append(depList, dep)
				}
			}
			if len(depList) > 0 {
				modMap[currentModuleName]["dependencies"] = depList
			}

		case line == "};":
			currentModuleName = ""
		}
	}

	names := make([]string, 0, len(modMap))
	for name := range modMap {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		modules = append(modules, modMap[name])
	}

	return modules
}

func decodeMOFText(content []byte) string {
	if len(content) == 0 {
		return ""
	}

	if len(content) >= 3 {
		if content[0] == 0xEF && content[1] == 0xBB && content[2] == 0xBF {
			return string(content[3:])
		}
	}

	if len(content) >= 2 {
		if content[0] == 0xFF && content[1] == 0xFE {
			return decodeUTF16(content[2:], true)
		}
		if content[0] == 0xFE && content[1] == 0xFF {
			return decodeUTF16(content[2:], false)
		}
	}

	if looksLikeUTF16LE(content) {
		return decodeUTF16(content, true)
	}
	if looksLikeUTF16BE(content) {
		return decodeUTF16(content, false)
	}

	if utf8.Valid(content) {
		return string(content)
	}

	return string(content)
}

func decodeUTF16(content []byte, littleEndian bool) string {
	if len(content) < 2 {
		return ""
	}
	if len(content)%2 != 0 {
		content = content[:len(content)-1]
	}

	u16 := make([]uint16, 0, len(content)/2)
	for i := 0; i+1 < len(content); i += 2 {
		if littleEndian {
			u16 = append(u16, binary.LittleEndian.Uint16(content[i:i+2]))
		} else {
			u16 = append(u16, binary.BigEndian.Uint16(content[i:i+2]))
		}
	}

	return string(utf16.Decode(u16))
}

func looksLikeUTF16LE(content []byte) bool {
	evenZeroRatio, oddZeroRatio, ok := zeroRatios(content)
	if !ok {
		return false
	}

	return oddZeroRatio > 0.30 && evenZeroRatio < 0.20
}

func looksLikeUTF16BE(content []byte) bool {
	evenZeroRatio, oddZeroRatio, ok := zeroRatios(content)
	if !ok {
		return false
	}

	return evenZeroRatio > 0.30 && oddZeroRatio < 0.20
}

func zeroRatios(content []byte) (float64, float64, bool) {
	sample := content
	if len(sample) > 256 {
		sample = sample[:256]
	}

	oddCount := 0
	oddZero := 0
	evenCount := 0
	evenZero := 0

	for i, b := range sample {
		if i%2 == 0 {
			evenCount++
			if b == 0 {
				evenZero++
			}
		} else {
			oddCount++
			if b == 0 {
				oddZero++
			}
		}
	}

	if oddCount == 0 || evenCount == 0 {
		return 0, 0, false
	}

	oddZeroRatio := float64(oddZero) / float64(oddCount)
	evenZeroRatio := float64(evenZero) / float64(evenCount)

	return evenZeroRatio, oddZeroRatio, true
}
