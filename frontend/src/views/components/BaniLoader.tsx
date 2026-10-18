const BANI = '/images/bani/bani-stand.webp'
const STEP_MS = 170

const BRICK_ROWS = [88, 73, 58]
const BRICK_COLUMNS = [12, 36, 60, 84]
const BRICK_TONES = ['#D9653B', '#C95A33', '#E0744A']
const BRICK_COUNT = BRICK_ROWS.length * BRICK_COLUMNS.length

function delay(step: number, extraMs = 0) {
  return { animationDelay: `${step * STEP_MS + extraMs}ms` }
}

// Animasi loading Bani: satu gambar Bani yang bergoyang pelan seperti sedang
// bekerja, di samping rumah yang digambar dengan kode dan dibangun bata demi
// bata, lalu atap, pintu, dan jendela. Semua gerakan dari CSS, jadi mulus dan
// ringan. Pengguna yang memilih mengurangi gerak melihat rumah yang sudah jadi.
function BuildingScene({ baniHeight, houseWidth }: { baniHeight: number; houseWidth: number }) {
  return (
    <div aria-hidden="true" className="flex items-end gap-1">
      <img
        src={BANI}
        alt=""
        draggable={false}
        style={{ height: baniHeight }}
        className="w-auto origin-bottom select-none motion-safe:animate-bani-work"
      />
      <House width={houseWidth} />
    </div>
  )
}

function House({ width }: { width: number }) {
  return (
    <svg viewBox="0 0 120 108" width={width} height={(width * 108) / 120} className="overflow-visible">
      <line x1="2" y1="103" x2="118" y2="103" stroke="#CBD5E1" strokeWidth="3" strokeLinecap="round" />
      {BRICK_ROWS.map((y, row) =>
        BRICK_COLUMNS.map((x, column) => (
          <rect
            key={`${row}-${column}`}
            x={x + (row % 2 === 1 ? 4 : 0)}
            y={y}
            width="22"
            height="13"
            rx="2.5"
            fill={BRICK_TONES[(row + column) % BRICK_TONES.length]}
            className="motion-safe:animate-brick"
            style={delay(row * BRICK_COLUMNS.length + column)}
          />
        )),
      )}
      <path
        d="M6 58 L60 22 L114 58 Z"
        fill="#0B2B6B"
        stroke="#0B2B6B"
        strokeWidth="6"
        strokeLinejoin="round"
        className="motion-safe:animate-brick"
        style={delay(BRICK_COUNT, 120)}
      />
      <rect x="52" y="78" width="18" height="25" rx="3" fill="#0B2B6B" className="motion-safe:animate-brick" style={delay(BRICK_COUNT + 1, 240)} />
      <rect x="22" y="63" width="18" height="14" rx="2.5" fill="#FFD60A" className="motion-safe:animate-brick" style={delay(BRICK_COUNT + 2, 240)} />
    </svg>
  )
}

// Layar penuh saat sesi diperiksa atau halaman pertama kali dimuat.
export function BaniScreen({ label }: { label: string }) {
  return (
    <div role="status" className="flex min-h-screen flex-col items-center justify-center gap-5 bg-slate-100 px-4">
      <BuildingScene baniHeight={148} houseWidth={124} />
      <p className="text-[13px] text-slate-500">{label}</p>
    </div>
  )
}

// Versi kecil di dalam kartu, muncul setelah 300 ms supaya data yang cepat
// datang tidak membuat kedipan.
export function BaniLoop({ label }: { label: string }) {
  return (
    <div role="status" className="flex items-center gap-3 py-3 motion-safe:animate-[fade_200ms_ease-out_300ms_both]">
      <BuildingScene baniHeight={48} houseWidth={40} />
      <span className="text-[13px] text-slate-500">{label}</span>
    </div>
  )
}
