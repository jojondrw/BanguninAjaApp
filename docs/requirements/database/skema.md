# Skema Database

Skema ini diturunkan dari livewire di `BanguninAja_livewire`, lalu dirapikan.
Livewire hanya gambaran tampilan, jadi bentuk datanya masih datar dan banyak
mengulang nama. Di sini semuanya dinormalkan.

## Aturan umum

| Hal | Keputusan |
|---|---|
| Kunci utama | UUID v4, dibuat di aplikasi lewat `entity.Base` |
| Waktu | `created_at` dan `updated_at` wajib, diisi otomatis |
| Uang | `bigint` dalam rupiah penuh. Tidak memakai float supaya tidak ada pembulatan aneh, dan rupiah tidak memakai sen |
| Jumlah barang | `numeric(14,2)` karena volume material bisa pecahan |
| Nama tabel | Bahasa Indonesia bentuk tunggal, contoh `proyek`, `rab_item`. Tabel milik auth memakai bahasa Inggris karena sudah terbentuk lebih dulu |
| Status | Disimpan sebagai `varchar` dan dijaga `CHECK`, bukan enum Postgres, supaya menambah pilihan baru tidak perlu mengubah tipe |

## Slice dan entity

| Slice | Entity |
|---|---|
| `master` | Wilayah, Satuan, Akun |
| `auth` | Peran, User, RefreshToken, PenggunaProyek |
| `proyek` | Proyek, TahapProyek, RabItem, DokumenIzin |
| `persediaan` | Material, Gudang, Stok, MutasiStok |
| `pengadaan` | Vendor, Permintaan, PermintaanItem, Pesanan, PesananItem, Penerimaan, PenerimaanItem |
| `aset` | Aset, Alat |
| `penjualan` | Pembeli, Unit, Prospek, Kontrak, Cicilan |
| `keuangan` | Anggaran, TransaksiKas, Jurnal, JurnalDetail |
| `tagihan` | Tagihan, Piutang, Hutang |
| `sdm` | Karyawan, Absensi, Gaji |
| `penilaian` | Dimensi, ProfilBangunan, Bobot |
| `lokasi` | Incaran, SkorDimensi, Perbandingan, PerbandinganItem |
| `laporan` | Laporan |

Totalnya 47 tabel, 66 foreign key, 67 check, 13 unique, dan 216 index.

## Yang ditambahkan karena livewire belum mencerminkan data asli

| Tambahan | Alasan |
|---|---|
| Slice `master` | Livewire menulis kota, satuan, dan nama akun sebagai teks bebas. Kalau dibiarkan, "Jakarta Pusat" dan "Jakarta pusat" jadi dua hal berbeda |
| `PermintaanItem`, `PesananItem`, `PenerimaanItem` | Livewire hanya menyimpan jumlah item sebagai angka, misalnya "12 item". Barang apa saja yang dipesan tidak tercatat, padahal itu inti pengadaan |
| `Material` dan `Gudang` | Halaman stok menulis nama material dan gudang sebagai teks. Stok harus melekat pada material dan gudang yang jelas |
| `Pembeli` | Livewire menulis nama pembeli di kontrak. Satu orang bisa punya beberapa kontrak, jadi datanya dipisah |
| `JurnalDetail` | Livewire menampilkan jurnal sebagai baris datar. Pembukuan berpasangan butuh satu jurnal berisi banyak baris debit dan kredit |
| `TahapProyek` | Tab jadwal di livewire hanya memakai tanggal mulai dan selesai proyek. Proyek konstruksi punya tahapan, dan progres dihitung per tahap |
| `Anggaran`, `TransaksiKas` | Livewire menghitung anggaran dari progres fisik dan menampilkan arus kas per bulan. Keduanya harus punya data sendiri |
| `ProfilBangunan`, `Dimensi`, `Bobot` | Tabel bobot penilaian di halaman pengaturan tadinya angka mati di layar |
| `Peran`, `PenggunaProyek` | Livewire menyimpan peran dan proyek sebagai teks pada pengguna. Satu orang bisa punya peran berbeda di proyek berbeda |

## Yang sengaja tidak dijadikan kolom

| Di livewire | Alasan |
|---|---|
| Umur piutang, misalnya "31 sampai 60 hari" | Dihitung dari `jatuh_tempo` dan tanggal hari ini. Kalau disimpan, besoknya sudah salah |
| Nilai buku aset | Selisih nilai perolehan dan akumulasi penyusutan |
| Progres proyek dalam persen | Tetap disimpan untuk sementara, tapi nanti diturunkan dari tahapan |
| Arus kas per bulan | Hasil penjumlahan `transaksi_kas`, bukan data mentah |
| Total RAB | Penjumlahan `rab_item` |

Aturannya: yang bisa dihitung ulang kapan saja tidak disimpan, supaya tidak ada
dua versi kebenaran.

## Kolom lokasi

`incaran.titik` bertipe `geometry(Point,4326)`. Nilainya ditulis lewat tipe
`geo.Point` di Go yang mengubah lintang bujur menjadi format yang dimengerti
PostGIS, dan membacanya kembali dari format biner PostGIS.

Pencarian berdasarkan area memakai index GiST `idx_incaran_titik`. Sudah
diperiksa dengan `EXPLAIN`, hasilnya Index Scan, bukan Seq Scan.

## Perilaku saat data induk dihapus

| Pola | Dipakai untuk |
|---|---|
| `CASCADE` | Anak yang tidak punya arti tanpa induknya, misalnya `rab_item` terhadap `proyek`, atau `cicilan` terhadap `kontrak` |
| `RESTRICT` | Data acuan yang dipakai catatan lain, misalnya `vendor`, `material`, `akun`. Menghapusnya akan ditolak selama masih dipakai |
| `SET NULL` | Kaitan yang boleh kosong, misalnya `alat.proyek_id` saat alatnya belum ditugaskan |

## Contoh aturan yang dijaga database

Ini beberapa yang sudah diuji dan benar-benar menolak data salah:

| Aturan | Constraint |
|---|---|
| Progres proyek 0 sampai 100 | `chk_proyek_progres` |
| Status proyek hanya lima pilihan | `chk_proyek_status` |
| Satu baris jurnal hanya boleh debit atau kredit, tidak dua duanya | `chk_jurnal_detail_satu_sisi` |
| Penyusutan tidak boleh melebihi nilai perolehan | `chk_aset_penyusutan_wajar` |
| Satu karyawan satu absensi per hari | `uq_absensi_karyawan_tanggal` |
| Stok tidak boleh minus | `chk_stok_jumlah` |
| Mutasi harus punya gudang asal atau tujuan | `chk_mutasi_stok_arah` |
| Pembayaran tidak boleh melebihi tagihan | `chk_tagihan_jumlah` |

## Menjalankan migrasi

```bash
docker compose exec backend go run ./cmd/migrate
```

Urutannya: pasang ekstensi, buat tabel per slice, lalu pasang index dan
constraint. Constraint dipasang setelah semua tabel ada, karena sebuah slice bisa
menunjuk tabel milik slice lain.

Semua perintahnya aman diulang. Index memakai `IF NOT EXISTS`, dan constraint
dibungkus pemeriksaan ke `pg_constraint` lebih dulu.
