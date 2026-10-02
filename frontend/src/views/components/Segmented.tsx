import type { Indicator } from './segmentIndicator'

export function SegmentIndicator({ indicator }: { indicator: Indicator | null }) {
  if (!indicator) {
    return null
  }

  return (
    <span
      aria-hidden="true"
      className="absolute top-[3px] bottom-[3px] left-0 rounded-[7px] bg-white shadow-hairline
                 motion-safe:transition-[transform,width] motion-safe:duration-300 motion-safe:ease-fluid"
      style={{ width: indicator.width, transform: `translateX(${indicator.left}px)` }}
    />
  )
}
