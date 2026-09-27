// Site Intelligence models. T16 only needs the picked point; the scoring types
// (predictive/descriptive) are wired in T17 against the T4 contract.

export interface LatLng {
  latitude: number
  longitude: number
}

// Default map view: centered on Indonesia so the whole archipelago is visible.
export const INDONESIA_CENTER: LatLng = { latitude: -2.5, longitude: 118.0 }
export const INDONESIA_ZOOM = 4.2
