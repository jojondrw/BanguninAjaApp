import { Layers } from 'lucide-react'
import { useEffect, useRef, useState } from 'react'
import {
  CircleMarker,
  Control,
  LatLngBounds,
  Map as LeafletMap,
  Popup,
  TileLayer,
  type Coords,
  type DoneCallback,
  type LeafletMouseEvent,
} from 'leaflet'
import 'leaflet/dist/leaflet.css'

import { getAccessToken } from '../../controllers/sessionStore'
import type { SavedLocation } from '../../models/erpApi'
import { API_BASE_URL } from '../../shared/apiClient'
import { CONTROL_CLASS } from './Form'
import { findMapLayer, MAP_LAYERS, type MapLayerId } from './mapLayers'

export interface MapPoint {
  latitude: number
  longitude: number
}

const INDONESIA_CENTER: [number, number] = [-2, 118]
const INDONESIA_ZOOM = 4.5
const FOCUS_ZOOM = 13
const FRAME_PADDING = 48
const FRAME_MAX_ZOOM = 11
const MAX_TILE_ZOOM = 19
const COORDINATE_DIGITS = 6

const PICKED_COLOR = '#fbbb16'
const SAVED_COLOR = '#0b2b6b'
const OVERLAY_MAX_ZOOM = 15
const OVERLAY_OPACITY = 0.85
const TILE_PREFIX = `${API_BASE_URL}/site/tiles/`
const OSM_TILES = 'https://tile.openstreetmap.org/{z}/{x}/{y}.png'
const OSM_ATTRIBUTION = '© <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'

// Lapisan data dari server butuh token, padahal ubin Leaflet biasanya hanya
// tag <img>. Ubin diambil lewat fetch berheader Authorization lalu dipasang
// sebagai blob, dan URL blob dilepas begitu ubinnya dibuang.
class AuthorizedTileLayer extends TileLayer {
  createTile(coords: Coords, done: DoneCallback): HTMLElement {
    const tile = document.createElement('img')
    tile.alt = ''
    tile.setAttribute('role', 'presentation')
    const accessToken = getAccessToken()
    fetch(this.getTileUrl(coords), { headers: accessToken ? { Authorization: `Bearer ${accessToken}` } : {} })
      .then((response) => {
        if (!response.ok) {
          throw new Error(`tile ${response.status}`)
        }
        return response.blob()
      })
      .then((blob) => {
        tile.onload = () => done(undefined, tile)
        tile.src = URL.createObjectURL(blob)
      })
      .catch((error: Error) => done(error, tile))
    return tile
  }
}

function releaseTile(event: { tile: HTMLElement }) {
  const source = (event.tile as HTMLImageElement).src
  if (source.startsWith('blob:')) {
    URL.revokeObjectURL(source)
  }
}

interface SiteMapProps {
  picked: MapPoint | null
  sites: SavedLocation[]
  focus: MapPoint | null
  onPick: (point: MapPoint) => void
  onOpenSite?: (id: string) => void
}

export function SiteMap({ picked, sites, focus, onPick, onOpenSite }: SiteMapProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const mapRef = useRef<LeafletMap | null>(null)
  const hasFramedSitesRef = useRef(false)
  const [layerId, setLayerId] = useState<MapLayerId | null>(null)
  const [zoom, setZoom] = useState(INDONESIA_ZOOM)
  const activeLayer = findMapLayer(layerId)
  const pickedLatitude = picked?.latitude ?? null
  const pickedLongitude = picked?.longitude ?? null

  useEffect(() => {
    const container = containerRef.current
    if (!container) {
      return
    }
    const map = new LeafletMap(container, {
      center: INDONESIA_CENTER,
      zoom: INDONESIA_ZOOM,
      zoomSnap: 0.25,
      zoomControl: false,
      preferCanvas: true,
    })
    new TileLayer(OSM_TILES, { maxZoom: MAX_TILE_ZOOM, attribution: OSM_ATTRIBUTION }).addTo(map)
    new Control.Zoom({ position: 'topright', zoomInTitle: 'Perbesar', zoomOutTitle: 'Perkecil' }).addTo(map)
    map.on('zoomend', () => setZoom(map.getZoom()))
    hasFramedSitesRef.current = false
    mapRef.current = map

    const observer = new ResizeObserver(() => map.invalidateSize())
    observer.observe(container)

    return () => {
      observer.disconnect()
      mapRef.current = null
      map.remove()
    }
  }, [])

  useEffect(() => {
    const map = mapRef.current
    if (!map || !activeLayer) {
      return
    }
    const overlay = new AuthorizedTileLayer(`${TILE_PREFIX}${activeLayer.id}/{z}/{x}/{y}`, {
      minZoom: activeLayer.minZoom,
      maxNativeZoom: OVERLAY_MAX_ZOOM,
      maxZoom: MAX_TILE_ZOOM,
      opacity: OVERLAY_OPACITY,
      attribution: activeLayer.attribution,
    })
    overlay.on('tileunload', releaseTile)
    overlay.addTo(map)

    return () => {
      overlay.remove()
    }
  }, [activeLayer])

  useEffect(() => {
    const map = mapRef.current
    if (!map) {
      return
    }
    function handleClick(event: LeafletMouseEvent) {
      onPick({
        latitude: roundCoordinate(event.latlng.lat),
        longitude: roundCoordinate(event.latlng.lng),
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
    const marker = new CircleMarker([pickedLatitude, pickedLongitude], {
      radius: 9,
      color: '#ffffff',
      weight: 3,
      fillColor: PICKED_COLOR,
      fillOpacity: 1,
      interactive: false,
    }).addTo(map)

    return () => {
      marker.remove()
    }
  }, [pickedLatitude, pickedLongitude])

  useEffect(() => {
    const map = mapRef.current
    if (!map) {
      return
    }
    const markers = sites.map((site) => savedSiteMarker(site, onOpenSite).addTo(map))
    if (!hasFramedSitesRef.current && sites.length > 0) {
      hasFramedSitesRef.current = true
      const bounds = new LatLngBounds(sites.map((site) => [site.latitude, site.longitude]))
      map.fitBounds(bounds, { padding: [FRAME_PADDING, FRAME_PADDING], maxZoom: FRAME_MAX_ZOOM, animate: false })
    }

    return () => {
      markers.forEach((marker) => marker.remove())
    }
  }, [sites, onOpenSite])

  useEffect(() => {
    const map = mapRef.current
    if (!map || !focus) {
      return
    }
    hasFramedSitesRef.current = true
    map.flyTo([focus.latitude, focus.longitude], Math.max(map.getZoom(), FOCUS_ZOOM))
  }, [focus])

  return (
    <div className="space-y-2">
      <div className="relative isolate">
        <div
          ref={containerRef}
          role="region"
          aria-label="Peta lokasi. Klik peta untuk memilih titik koordinat."
          className="h-80 w-full overflow-hidden rounded-xl bg-slate-100 shadow-hairline sm:h-96"
        />
        <div className="absolute top-2.5 left-2.5 z-[1000] flex items-center gap-2 rounded-xl bg-white/90 py-1 pr-1 pl-2.5 shadow-panel backdrop-blur">
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

function roundCoordinate(value: number): number {
  return Number(value.toFixed(COORDINATE_DIGITS))
}

function savedSiteMarker(site: SavedLocation, onOpenSite?: (id: string) => void): CircleMarker {
  const marker = new CircleMarker([site.latitude, site.longitude], {
    radius: 7,
    color: '#ffffff',
    weight: 2,
    fillColor: SAVED_COLOR,
    fillOpacity: 1,
    bubblingMouseEvents: false,
  })
  const popup = new Popup({ offset: [0, -4] })
  popup.setContent(
    savedSitePopup(
      site,
      onOpenSite
        ? () => {
            marker.closePopup()
            onOpenSite(site.id)
          }
        : undefined,
    ),
  )
  return marker.bindPopup(popup)
}

function savedSitePopup(site: SavedLocation, onOpen?: () => void): HTMLElement {
  const content = document.createElement('div')
  const name = document.createElement('p')
  const score = document.createElement('p')

  name.className = 'text-sm font-semibold text-slate-900'
  name.textContent = site.name
  score.className = 'mt-0.5 text-xs text-slate-600'
  score.textContent = `Skor kelayakan ${site.score}`
  content.append(name, score)
  if (onOpen) {
    const open = document.createElement('button')
    open.type = 'button'
    open.className = 'mt-2 rounded-md bg-navy-600 px-2.5 py-1 text-xs font-medium text-white hover:bg-navy-700'
    open.textContent = 'Lihat rincian'
    open.addEventListener('click', onOpen)
    content.append(open)
  }
  return content
}
