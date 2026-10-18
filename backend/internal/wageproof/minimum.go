package wageproof

type MinimumWage struct {
	Amount int64
	Source string
}

var minimumWages = map[string]map[int]MinimumWage{
	"3273": {2026: {4_737_678, "UMK Kota Bandung 2026, Kepgub Jabar 561/Kep.862-Kesra/2025"}},
	"3274": {2026: {2_878_646, "UMK Kota Cirebon 2026, Kepgub Jabar 561/Kep.862-Kesra/2025 (satu sumber, cocokkan dengan salinan SK)"}},
	"3276": {2026: {5_552_662, "UMK Kota Depok 2026, Kepgub Jabar 561/Kep.862-Kesra/2025 (sumber berbeda 5.522.662 dan 5.552.662, dipakai yang lebih tinggi)"}},
	"7371": {2026: {4_148_719, "UMK Kota Makassar 2026, SK Gubernur Sulsel 2142/XII/2025 (sumber berbeda 4.148.179 dan 4.148.719, dipakai yang lebih tinggi)"}},
}

func MinimumWageFor(regionCode string, year int) (MinimumWage, bool) {
	wage, ok := minimumWages[regionCode][year]
	return wage, ok
}
