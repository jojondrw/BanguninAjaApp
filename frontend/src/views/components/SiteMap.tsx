import maplibregl from 'maplibre-gl'
import type { StyleSpecification } from 'maplibre-gl'
import { useEffect, useRef } from 'react'
import 'maplibre-gl/dist/maplibre-gl.css'

import { INDONESIA_CENTER, INDONESIA_ZOOM, type LatLng } from '../../models/site'

// A no-API-key basemap built from OpenStreetMap's raster tiles. Kept inline so the
// map works out of the box; swap this for a team tile provider later if desired.
const OSM_STYLE: StyleSpecification = {
  version: 8,
  sources: {
    osm: {
      type: 'raster',
      tiles: ['https://tile.openstreetmap.org/{z}/{x}/{y}.png'],
      tileSize: 256,
      attribution: '&copy; OpenStreetMap contributors',
    },
  },
  layers: [{ id: 'osm', type: 'raster', source: 'osm' }],
}

interface SiteMapProps {
  // Called whenever the user clicks the map to pick a point.
  onPick?: (point: LatLng) => void
  // Currently selected point, rendered as a marker. Controlled by the parent.
  selected?: LatLng | null
}

export function SiteMap({ onPick, selected }: SiteMapProps) {
  const containerRef = useRef<HTMLDivElement | null>(null)
  const mapRef = useRef<maplibregl.Map | null>(null)
  const markerRef = useRef<maplibregl.Marker | null>(null)
  // Keep the latest onPick in a ref so the map is created once, not per render.
  const onPickRef = useRef(onPick)
  useEffect(() => {
    onPickRef.current = onPick
  }, [onPick])

  // Create the map exactly once.
  useEffect(() => {
    if (!containerRef.current || mapRef.current) {
      return
    }

    const map = new maplibregl.Map({
      container: containerRef.current,
      style: OSM_STYLE,
      center: [INDONESIA_CENTER.longitude, INDONESIA_CENTER.latitude],
      zoom: INDONESIA_ZOOM,
      attributionControl: { compact: true },
    })
    map.addControl(new maplibregl.NavigationControl({ showCompass: false }), 'top-right')
    map.on('click', (event) => {
      onPickRef.current?.({ latitude: event.lngLat.lat, longitude: event.lngLat.lng })
    })

    mapRef.current = map
    return () => {
      map.remove()
      mapRef.current = null
      markerRef.current = null
    }
  }, [])

  // Reflect the selected point as a marker whenever it changes.
  useEffect(() => {
    const map = mapRef.current
    if (!map) {
      return
    }

    if (!selected) {
      markerRef.current?.remove()
      markerRef.current = null
      return
    }

    const position: [number, number] = [selected.longitude, selected.latitude]
    if (markerRef.current) {
      markerRef.current.setLngLat(position)
    } else {
      markerRef.current = new maplibregl.Marker({ color: '#0b2b6b' }).setLngLat(position).addTo(map)
    }
  }, [selected])

  return (
    <div
      ref={containerRef}
      role="application"
      aria-label="Peta pemilihan lokasi"
      className="h-[520px] w-full overflow-hidden rounded-xl border border-slate-200"
    />
  )
}
