import { Layers } from 'lucide-react'
import { useEffect, useRef, useState } from 'react'
import {
  MapLibreMap,
  Marker,
  NavigationControl,
  Popup,
  setWorkerUrl,
  type MapMouseEvent,
  type MapOptions,
} from 'maplibre-gl'
import maplibreWorkerUrl from 'maplibre-gl/dist/maplibre-gl-worker.mjs?worker&url'
import 'maplibre-gl/dist/maplibre-gl.css'

import { getAccessToken } from '../../controllers/sessionStore'
import type { SavedLocation } from '../../models/erpApi'
import { API_BASE_URL } from '../../shared/apiClient'
import { CONTROL_CLASS } from './Form'
import { findMapLayer, MAP_LAYERS, type MapLayerId } from './mapLayers'

setWorkerUrl(maplibreWorkerUrl)

export interface MapPoint {
  latitude: number
  longitude: number
}

const INDONESIA_CENTER: [number, number] = [118, -2]
const INDONESIA_ZOOM = 4.3
const FOCUS_ZOOM = 13
const MAX_TILE_ZOOM = 19
const COORDINATE_DIGITS = 6

const PICKED_COLOR = '#fbbb16'
const SAVED_COLOR = '#0b2b6b'
const NON_PICKING_TARGETS = '.maplibregl-marker, .maplibregl-popup'
const OVERLAY_ID = 'lapisan-data'
const OVERLAY_MAX_ZOOM = 15
const TILE_PREFIX = `${API_BASE_URL}/site/tiles/`

const OSM_STYLE: MapOptions['style'] = {
  version: 8,
  sources: {
    osm: {
      type: 'raster',
      tiles: ['https://tile.openstreetmap.org/{z}/{x}/{y}.png'],
      tileSize: 256,
      maxzoom: MAX_TILE_ZOOM,
      attribution: '© <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
    },
  },
  layers: [{ id: 'osm', type: 'raster', source: 'osm' }],
}

interface SiteMapProps {
  picked: MapPoint | null
  sites: SavedLocation[]
  focus: MapPoint | null
  onPick: (point: MapPoint) => void
}

export function SiteMap({ picked, sites, focus, onPick }: SiteMapProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const mapRef = useRef<MapLibreMap | null>(null)
  const [isMapReady, setIsMapReady] = useState(false)
  const [layerId, setLayerId] = useState<MapLayerId | null>(null)
  const [zoom, setZoom] = useState(INDONESIA_ZOOM)
  const activeLayer = findMapLayer(layerId)
  const pickedLatitude = picked?.latitude ?? null
  const pickedLongitude = picked?.longitude ?? null

  useEffect(() => {
    if (!containerRef.current) {
      return
    }
    const map = new MapLibreMap({
      container: containerRef.current,
      style: OSM_STYLE,
      center: INDONESIA_CENTER,
      zoom: INDONESIA_ZOOM,
      transformRequest: (url) => {
        const accessToken = getAccessToken()
        return url.startsWith(TILE_PREFIX) && accessToken
          ? { url, headers: { Authorization: `Bearer ${accessToken}` } }
          : { url }
      },
    })
    map.addControl(new NavigationControl({ showCompass: false }), 'top-right')
    map.on('load', () => setIsMapReady(true))
    map.on('zoomend', () => setZoom(map.getZoom()))
    mapRef.current = map

    return () => {
      mapRef.current = null
      map.remove()
    }
  }, [])

  useEffect(() => {
    const map = mapRef.current
    if (!map || !isMapReady || !activeLayer) {
      return
    }
    map.addSource(OVERLAY_ID, {
      type: 'raster',
      tiles: [`${TILE_PREFIX}${activeLayer.id}/{z}/{x}/{y}`],
      tileSize: 256,
      minzoom: activeLayer.minZoom,
      maxzoom: OVERLAY_MAX_ZOOM,
      attribution: activeLayer.attribution,
    })
    map.addLayer({ id: OVERLAY_ID, type: 'raster', source: OVERLAY_ID, paint: { 'raster-opacity': 0.85 } })

    return () => {
      if (map.getLayer(OVERLAY_ID)) {
        map.removeLayer(OVERLAY_ID)
      }
      if (map.getSource(OVERLAY_ID)) {
        map.removeSource(OVERLAY_ID)
      }
    }
  }, [activeLayer, isMapReady])

  useEffect(() => {
    const map = mapRef.current
    if (!map) {
      return
    }
    function handleClick(event: MapMouseEvent) {
      if (isOverlayClick(event)) {
        return
      }
      onPick({
        latitude: roundCoordinate(event.lngLat.lat),
        longitude: roundCoordinate(event.lngLat.lng),
      })
    }
    map.on('click', handleClick)

    return () => {
      map.off('click', handleClick)
    }
  }, [onPick])

  useEffect(() => {
    const map = mapRef.current
    if (!map || pickedLatitude === null || pickedLongitude === null) {
      return
    }
    const marker = new Marker({ color: PICKED_COLOR })
      .setLngLat([pickedLongitude, pickedLatitude])
      .addTo(map)

    return () => {
      marker.remove()
    }
  }, [pickedLatitude, pickedLongitude])

  useEffect(() => {
    const map = mapRef.current
    if (!map) {
      return
    }
    const markers = sites.map((site) => savedSiteMarker(site).addTo(map))

    return () => {
      markers.forEach((marker) => marker.remove())
    }
  }, [sites])

  useEffect(() => {
    const map = mapRef.current
    if (!map || !focus) {
      return
    }
    map.flyTo({
      center: [focus.longitude, focus.latitude],
      zoom: Math.max(map.getZoom(), FOCUS_ZOOM),
    })
  }, [focus])

  return (
    <div className="space-y-2">
      <div className="relative">
        <div
          ref={containerRef}
          role="region"
          aria-label="Peta lokasi. Klik peta untuk memilih titik koordinat."
          className="h-80 w-full overflow-hidden rounded-xl shadow-hairline sm:h-96"
        />
        <div className="absolute top-2.5 left-2.5 flex items-center gap-2 rounded-xl bg-white/90 py-1 pr-1 pl-2.5 shadow-panel backdrop-blur">
          <Layers aria-hidden="true" className="size-4 text-slate-500" strokeWidth={1.8} />
          <label htmlFor="lapisan-peta" className="sr-only">
            Lapisan data di peta
          </label>
          <select
            id="lapisan-peta"
            value={layerId ?? ''}
            onChange={(event) => setLayerId((event.target.value || null) as MapLayerId | null)}
            className={`${CONTROL_CLASS} h-8 max-w-52 px-2 text-[13px]`}
          >
            <option value="">Tanpa lapisan data</option>
            {MAP_LAYERS.map((layer) => (
              <option key={layer.id} value={layer.id}>
                {layer.label}
              </option>
            ))}
          </select>
        </div>
      </div>
      {activeLayer ? (
        <div className="rounded-xl bg-slate-50 px-3 py-2.5">
          <div className="flex flex-wrap items-baseline justify-between gap-x-3 gap-y-1">
            <p className="text-xs font-medium text-slate-700">{activeLayer.description}</p>
            {zoom < activeLayer.minZoom ? (
              <p role="status" className="text-xs text-amber-700">
                Perbesar peta untuk melihat lapisan ini (mulai zoom {activeLayer.minZoom}).
              </p>
            ) : null}
          </div>
          <div aria-hidden="true" className="mt-2 h-2 rounded-full" style={{ background: activeLayer.gradient }} />
          <div className="mt-1 flex justify-between text-[11px] text-slate-500">
            {activeLayer.ticks.map((tick) => (
              <span key={tick}>{tick}</span>
            ))}
          </div>
        </div>
      ) : null}
      <div className="flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-slate-600">
        <LegendDot color={PICKED_COLOR} label="Titik yang dipilih" />
        <LegendDot color={SAVED_COLOR} label="Lokasi yang sudah dievaluasi, klik untuk melihat skor" />
      </div>
    </div>
  )
}

function LegendDot({ color, label }: { color: string; label: string }) {
  return (
    <span className="inline-flex items-center gap-1.5">
      <span aria-hidden="true" className="h-2.5 w-2.5 rounded-full" style={{ backgroundColor: color }} />
      {label}
    </span>
  )
}

function isOverlayClick(event: MapMouseEvent): boolean {
  const target = event.originalEvent.target
  return target instanceof Element && target.closest(NON_PICKING_TARGETS) !== null
}

function roundCoordinate(value: number): number {
  return Number(value.toFixed(COORDINATE_DIGITS))
}

function savedSiteMarker(site: SavedLocation): Marker {
  return new Marker({ color: SAVED_COLOR, scale: 0.8 })
    .setLngLat([site.longitude, site.latitude])
    .setPopup(new Popup({ offset: 24 }).setDOMContent(savedSitePopup(site)))
}

function savedSitePopup(site: SavedLocation): HTMLElement {
  const content = document.createElement('div')
  const name = document.createElement('p')
  const score = document.createElement('p')

  name.className = 'text-sm font-semibold text-slate-900'
  name.textContent = site.name
  score.className = 'mt-0.5 text-xs text-slate-600'
  score.textContent = `Skor kelayakan ${site.score}`
  content.append(name, score)
  return content
}
