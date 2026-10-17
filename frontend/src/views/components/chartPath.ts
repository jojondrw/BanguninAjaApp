export type Point = readonly [number, number]

// Kurva monoton (Fritsch-Carlson): melengkung halus seperti grafik Apple, tetapi
// tidak pernah melewati titik data, jadi kurva S tidak tampak turun lalu naik.
export function smoothPath(points: readonly Point[]): string {
  if (points.length === 0) {
    return ''
  }
  if (points.length === 1) {
    return `M${points[0][0]},${points[0][1]}`
  }

  const count = points.length
  const slopes: number[] = []
  for (let index = 0; index < count - 1; index++) {
    const [x0, y0] = points[index]
    const [x1, y1] = points[index + 1]
    slopes.push(x1 === x0 ? 0 : (y1 - y0) / (x1 - x0))
  }

  const tangents = points.map((_, index) => {
    if (index === 0) {
      return slopes[0]
    }
    if (index === count - 1) {
      return slopes[count - 2]
    }
    const before = slopes[index - 1]
    const after = slopes[index]
    return before * after <= 0 ? 0 : (before + after) / 2
  })

  for (let index = 0; index < count - 1; index++) {
    const slope = slopes[index]
    if (slope === 0) {
      tangents[index] = 0
      tangents[index + 1] = 0
      continue
    }
    const alpha = tangents[index] / slope
    const beta = tangents[index + 1] / slope
    const length = Math.hypot(alpha, beta)
    if (length > 3) {
      const factor = 3 / length
      tangents[index] = factor * alpha * slope
      tangents[index + 1] = factor * beta * slope
    }
  }

  let path = `M${points[0][0]},${points[0][1]}`
  for (let index = 0; index < count - 1; index++) {
    const [x0, y0] = points[index]
    const [x1, y1] = points[index + 1]
    const third = (x1 - x0) / 3
    path += ` C${x0 + third},${y0 + tangents[index] * third} ${x1 - third},${y1 - tangents[index + 1] * third} ${x1},${y1}`
  }
  return path
}

// Area di bawah kurva sampai garis dasar, untuk isian gradasi yang memudar.
export function areaPath(points: readonly Point[], baseline: number): string {
  if (points.length < 2) {
    return ''
  }
  const first = points[0]
  const last = points[points.length - 1]
  return `${smoothPath(points)} L${last[0]},${baseline} L${first[0]},${baseline} Z`
}

// Batang dengan sudut ujung luar sedikit membulat dan sisi di garis nol tetap
// rata. Jari-jari dibatasi tinggi batang, jadi batang pendek tetap persegi.
export function barPath(x: number, top: number, width: number, height: number, direction: 'up' | 'down', radius = 3): string {
  if (width <= 0 || height <= 0) {
    return ''
  }
  const r = Math.min(radius, width / 2, height)
  const right = x + width
  const bottom = top + height
  if (direction === 'up') {
    return `M${x},${bottom} L${x},${top + r} A${r},${r} 0 0 1 ${x + r},${top} L${right - r},${top} A${r},${r} 0 0 1 ${right},${top + r} L${right},${bottom} Z`
  }
  return `M${x},${top} L${x},${bottom - r} A${r},${r} 0 0 0 ${x + r},${bottom} L${right - r},${bottom} A${r},${r} 0 0 0 ${right},${bottom - r} L${right},${top} Z`
}
