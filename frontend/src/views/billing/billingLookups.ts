import { vendorOptions } from '../../models/lookupApi'

// Faktur dan utang lama bisa menunjuk vendor yang sudah nonaktif, jadi
// pencarian vendor di halaman tagihan tidak menyaring status aktif.
export const allVendorOptions = vendorOptions()
