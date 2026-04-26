package handlers

import (
	"encoding/binary"
	"testing"
	"unicode/utf16"
)

func TestParseModulesFromMOFBytes_UTF16LE(t *testing.T) {
	mof := `instance of cChocoSource as $cChocoSource1ref
{
 ModuleName = "cChoco";
 ModuleVersion = "2.6.0.0";
};
instance of MSFT_RoleResource as $MSFT_RoleResource1ref
{
 ModuleName = "PSDesiredStateConfiguration";
 ModuleVersion = "1.0";
};`

	modules := parseModulesFromMOFBytes(utf16LEWithBOM(mof))

	if len(modules) != 2 {
		t.Fatalf("expected 2 modules, got %d", len(modules))
	}

	cChoco := findModuleByName(modules, "cChoco")
	if cChoco == nil {
		t.Fatalf("expected module cChoco to be present")
	}
	if cChoco["version"] != "2.6.0.0" {
		t.Fatalf("expected cChoco version 2.6.0.0, got %v", cChoco["version"])
	}

	psDsc := findModuleByName(modules, "PSDesiredStateConfiguration")
	if psDsc == nil {
		t.Fatalf("expected module PSDesiredStateConfiguration to be present")
	}
	if psDsc["version"] != "1.0" {
		t.Fatalf("expected PSDesiredStateConfiguration version 1.0, got %v", psDsc["version"])
	}
}

func TestParseModulesFromMOFBytes_UTF8AndDuplicateModuleName(t *testing.T) {
	mof := `instance of cChocoPackageInstall as $r1
{
 ModuleName = "cChoco";
 ModuleVersion = "2.6.0.0";
};
instance of cChocoPackageInstall as $r2
{
 ModuleName = "cChoco";
 ModuleVersion = "2.6.0.0";
};`

	modules := parseModulesFromMOFBytes([]byte(mof))

	if len(modules) != 1 {
		t.Fatalf("expected 1 unique module, got %d", len(modules))
	}
	if modules[0]["name"] != "cChoco" {
		t.Fatalf("expected module name cChoco, got %v", modules[0]["name"])
	}
	if modules[0]["version"] != "2.6.0.0" {
		t.Fatalf("expected module version 2.6.0.0, got %v", modules[0]["version"])
	}
}

func TestParseModulesFromMOFBytes_UTF8WithBOM(t *testing.T) {
	mof := `instance of cChocoSource as $r1
{
 ModuleName = "cChoco";
 ModuleVersion = "2.6.0.0";
};`

	modules := parseModulesFromMOFBytes(utf8WithBOM(mof))

	if len(modules) != 1 {
		t.Fatalf("expected 1 module, got %d", len(modules))
	}
	if modules[0]["name"] != "cChoco" {
		t.Fatalf("expected module name cChoco, got %v", modules[0]["name"])
	}
}

func TestParseModulesFromMOFBytes_UTF16BEWithoutBOM(t *testing.T) {
	mof := `instance of cChocoSource as $r1
{
 ModuleName = "cChoco";
 ModuleVersion = "2.6.0.0";
};`

	modules := parseModulesFromMOFBytes(utf16BEWithoutBOM(mof))

	if len(modules) != 1 {
		t.Fatalf("expected 1 module, got %d", len(modules))
	}
	if modules[0]["name"] != "cChoco" {
		t.Fatalf("expected module name cChoco, got %v", modules[0]["name"])
	}
	if modules[0]["version"] != "2.6.0.0" {
		t.Fatalf("expected module version 2.6.0.0, got %v", modules[0]["version"])
	}
}

func utf16LEWithBOM(s string) []byte {
	u16 := utf16.Encode([]rune(s))
	out := make([]byte, 2+len(u16)*2)
	out[0] = 0xFF
	out[1] = 0xFE
	for i, r := range u16 {
		binary.LittleEndian.PutUint16(out[2+i*2:2+i*2+2], r)
	}
	return out
}

func utf16BEWithoutBOM(s string) []byte {
	u16 := utf16.Encode([]rune(s))
	out := make([]byte, len(u16)*2)
	for i, r := range u16 {
		binary.BigEndian.PutUint16(out[i*2:i*2+2], r)
	}
	return out
}

func utf8WithBOM(s string) []byte {
	out := make([]byte, 0, len(s)+3)
	out = append(out, 0xEF, 0xBB, 0xBF)
	out = append(out, []byte(s)...)
	return out
}

func findModuleByName(modules []map[string]interface{}, name string) map[string]interface{} {
	for _, m := range modules {
		if m["name"] == name {
			return m
		}
	}
	return nil
}
