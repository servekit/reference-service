package data

import (
	"regexp"
	"testing"
)

func TestLanguagesIntegrity(t *testing.T) {
	if n := len(Languages); n < 150 || n > 400 {
		t.Fatalf("language count %d outside [150,400]", n)
	}
	tagRe := regexp.MustCompile(`^[a-z]{2}(-Hant|-Hans)?$`)
	for tag, l := range Languages {
		if l.Tag != tag {
			t.Errorf("key %q != tag %q", tag, l.Tag)
		}
		if !tagRe.MatchString(tag) {
			t.Errorf("bad tag %q", tag)
		}
	}
	for _, want := range []string{"zh-Hans", "zh-Hant", "en", "ja", "ko", "fr", "de", "pt", "es", "ar", "ru"} {
		if Languages[want].Tag == "" {
			t.Errorf("expected language %s missing", want)
		}
	}
	for _, l := range Locales {
		if len(LanguageNames[l]) != len(Languages) {
			t.Fatalf("locale %s: %d names for %d languages", l, len(LanguageNames[l]), len(Languages))
		}
		for tag, n := range LanguageNames[l] {
			if n == "" {
				t.Errorf("locale %s: empty name for %s", l, tag)
			}
		}
	}
}

func TestLanguagesNativeNames(t *testing.T) {
	cases := map[string]string{
		"zh-Hans": "简体中文",
		"zh-Hant": "繁體中文",
		"ja":      "日本語",
		"ko":      "한국어",
		"en":      "English",
		"ar":      "العربية",
		"ru":      "русский",
		"th":      "ไทย",
	}
	for tag, want := range cases {
		if got := Languages[tag].NativeName; got != want {
			t.Errorf("%s native = %q, want %q", tag, got, want)
		}
	}
}

func TestLanguagesGoldenRows(t *testing.T) {
	if got := LanguageNames["zh-Hans"]["en"]; got != "英语" {
		t.Errorf("zh-Hans en = %q, want 英语", got)
	}
	if got := LanguageNames["en"]["zh-Hans"]; got != "Simplified Chinese" {
		t.Errorf("en zh-Hans = %q, want Simplified Chinese", got)
	}
	if got := LanguageNames["ja"]["zh-Hans"]; got != "簡体中国語" {
		t.Errorf("ja zh-Hans = %q, want 簡体中国語", got)
	}
}
