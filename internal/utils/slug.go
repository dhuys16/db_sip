package utils

import (
	"regexp"
	"strings"
)

// GenerateSlug mengubah string judul menjadi slug URL-friendly.
// Contoh: "Bagi Sembako Ramadhan 2026!" -> "bagi-sembako-ramadhan-2026"
func GenerateSlug(title string) string {
	slug := strings.ToLower(strings.TrimSpace(title))

	// Buang karakter selain huruf, angka, spasi, dan dash
	reNonAlnum := regexp.MustCompile(`[^a-z0-9\s-]`)
	slug = reNonAlnum.ReplaceAllString(slug, "")

	// Ganti spasi/underscore jadi dash
	reSpace := regexp.MustCompile(`[\s_]+`)
	slug = reSpace.ReplaceAllString(slug, "-")

	// Rapikan dash berturut-turut
	reDash := regexp.MustCompile(`-+`)
	slug = reDash.ReplaceAllString(slug, "-")

	return strings.Trim(slug, "-")
}