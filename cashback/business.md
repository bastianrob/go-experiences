# In the World of Cashback Pt.2

Post ini berisi **sudut pandang** gw terhadap dunia per-cashback-an.

> It is not necessarily true, just my honest opinion based on observation.

## Alur bisnis

Ada at least 3 pihak yang terlibat: (but not always. It can be more than 3! we'll discuss this later below)

* `Platform` adalah penyedia infrastruktur cashback. e.g: `Go***`, `O*O`
* `Merchant` adalah pihak pengusaha yang menyediakan produk/jasa dan bekerja sama dengan penyedia `Platform`
* `User` adalah pengguna `Platform` yang berbelanja / menggunakan service dari `Merchant`

Sebagai `User`:

* Gw belanja ke `Merchant`
* Gw melakukan pembayaran menggunakan `Platform`
* `Platform` menjanjikan cashback sebesar `N%` dari total nilai transaksi `T` dengan
  * Minimal nilai transaksi adalah `X`, dan
  * Maximum nilai cashback adalah `Y`
* Gw tekan konfirmasi pembayaran di `Platform`
* `Platform` mengembalikan saldo gw sebesar `Z` rupiah

Question time:
Berapa nilai uang yang diterima oleh `Merchant`?

* `Money = T - Z`
* `Money = T`

My take on this: bisa dua-duanya, ato bahkan ada opsi ke-tiga

* `Money = T - Y` dimana `Y` adalah sesuatu yang di kenal dengan `MDR`

## Ilusi Subsidi dan Bakar Uang

`MRD` Adalah `fee` yang dibebankan ke pihak `Merchant` setiap kali terjadi transaksi melalui penyedia `Platform`.
Fee ini berupa potongan terhadap setiap transaksi, di mana sebagai `Merchant`:

* Gw bekerja sama dengan penyedia `Platform`
* Gw dan `Platform` menjalin kesepakatan tentang besaran `MDR` yang di bebankan
* Pada saat terjadi transaksi, `Platform` menerima uang dari `User` sejumlah nilai transaksi yang terjadi
* Dari total nilai transaksi `T`, `Platform` akan memotong `T` sebesar nilai `MDR` yang telah di sepakati
* Terakhir, `Platform` akan memasukkan `T - MDR` sebagai deposit ke dalam akun yang dimiliki `Merchant`

Mari men-simulasikan suatu skenario dimana `MDR` nya `10%`

> `Platform` menyediakan cashback `10%` pada setiap transaksi di `Merchant` manapun, dengan peraturan:

* Minimal transaksi = Rp. 50.000,-
* Maximal cashback = Rp. 10.000,-

### `User` melakukan transaksi di `Merchant` dengan total nilai transaksi = Rp. 150.000,-

* Maka `Fee` yang dibebankan ke `Merchant` adalah Rp. 150.000,- * 10% = Rp. 15.000,-
* Dengan demikian, `Platform` menerima total uang sebesar Rp. 150.000,-
* Dimana Rp. 10.000,- akan diberikan ke `User`
* Lalu Rp. 135.000,- akan diberikan ke `Merchant`
* Dan Rp. 5.000,- akan diterima oleh `Platform` sebagai profit margin

### `User` melakukan transaksi di `Merchant` dengan total nilai transaksi = Rp. 100.000,-

* Maka `Fee` yang dibebankan ke `Merchant` adalah Rp. 100.000,- * 10% = Rp. 10.000,-
* Dengan demikian, `Platform` menerima total uang sebesar Rp. 100.000,-
* Dimana Rp. 10.000,- akan diberikan ke `User`
* Lalu Rp. 90.000,- akan diberikan ke `Merchant`
* `Platform` tidak menerima profit margin

### `User` melakukan transaksi di `Merchant` dengan total nilai transaksi = Rp. 50.000,-

* Maka `Fee` yang dibebankan ke `Merchant` adalah Rp. 50.000,- * 10% = Rp. 50.000,-
* Dengan demikian, `Platform` menerima total uang sebesar Rp. 50.000,-
* Dimana Rp. 5.000,- akan diberikan ke `User`
* Lalu Rp. 45.000,- akan diberikan ke `Merchant`
* `Platform` tidak menerima profit margin

### `User` melakukan transaksi di `Merchant` dengan total nilai transaksi = Rp. 49.000,-

* Maka `Fee` yang dibebankan ke `Merchant` adalah Rp. 49.000,- * 10% = Rp. 4.900,-
* Dengan demikian, `Platform` menerima total uang sebesar Rp. 49.000,-
* Dimana Rp. 44.100,- akan diberikan ke `Merchant`
* Dan Rp. 4.900,- akan diterima oleh `Platform` sebagai profit margin

Notice dimana seluruh skenario yang kita simulasikan diatas, gak 1 kali pun `Platform` menerima kerugian?
Mentok-mentok juga `Platform` cuma terima impas di range harga tertentu.

Maximal cashback dan minimal transaksi lah yang menyelamatkan platform dari bakar uang.

## Subsidi Silang

Kalo kalian perhati-in, saat ini ada 2 tipe `Platform` yang berkeliaran:

* `e-money` kayak `O*O` dan `Go***`
* `e-wallet` kayak `D***`

> Bedanya apa ya bro?

Jadi di `e-money` uang anda ada dipegang `Platform` dan anda harus `top-up`
Kalo `e-wallet`, layaknya dompet biasa, kita bisa `link` kartu kredit ato debit (VISA/Mastercard) kita ke penyedia `Platform`

Menurut gw 2 tipe `Platform` ini lama kelamaan akan converge ke 1 layanan terpadu dimana kita bisa `top-up` dan `link` kartu di dalam 1 `Platform` yang sama.
Kenapa? Karena ada skema subsidi silang!
Masih ingat tadi di-atas gw bilang gak selalu cuma ad 3 pihak yang terlibat dalam per-cashback-an ini?

**Please welcome `Bank` as the 4th player in this party...**

Jadi `Bank` itu rupanya **bisa aja** punya program `pencitraan`.
Penyedia `Platform` bisa, dan gw rasa dengan mudah, berkolaborasi dengan pihak `Bank` untuk melakukan subsidi silang, dimana:

* Pihak `Platform` akan menyediakan promo khusus
* Seluruh / sebagian biaya cashback dalam promo ybs, akan ditanggung oleh pihak `Bank`
* Jika dan hanya jika pembayaran terjadi menggunakan kartu yang diterbitkan oleh `Bank` partner
* Dan `Bank` akan meminta exposure dalam bentuk banner / iklan promo yang dilakukan oleh pihak `Platform`

Mari kita lanjutkan simulasi skenario dimana

* `MDR` merchant `10%`
* Promo cashback `50%` ditanggung seluruhnya oleh pihak `Bank`
* Pembayaran menggunakan kartu yang diterbitkan oleh pihak `Bank`
* **Tanpa** Minimal transaksi!
* Maximal cashback = Rp. 50.000,-

### `User` melakukan transaksi di `Merchant` dengan total nilai transaksi = Rp. 150.000,- menggunakan kartu `Bank` partner

* Maka `Fee` yang dibebankan ke `Merchant` adalah Rp. 150.000,- * 10% = Rp. 15.000,-
* Dengan demikian, `Platform` menerima total uang sebesar Rp. 150.000,-
* Dimana Rp. 50.000,- akan diberikan ke `User` dan ditanggung oleh pihak `Bank`
* Lalu Rp. 135.000,- akan diberikan ke `Merchant`
* Dan Rp. 15.000,- akan diterima oleh `Platform` sebagai profit margin

### `User` melakukan transaksi di `Merchant` dengan total nilai transaksi = Rp. 100.000,- menggunakan kartu `Bank` partner

* Maka `Fee` yang dibebankan ke `Merchant` adalah Rp. 100.000,- * 10% = Rp. 10.000,-
* Dengan demikian, `Platform` menerima total uang sebesar Rp. 100.000,-
* Dimana Rp. 50.000,- akan diberikan ke `User` dan ditanggung oleh pihak `Bank`
* Lalu Rp. 90.000,- akan diberikan ke `Merchant`
* Dan Rp. 10.000,- akan diterima oleh `Platform` sebagai profit margin

### `User` melakukan transaksi di `Merchant` dengan total nilai transaksi = Rp. 50.000,- menggunakan kartu `Bank` partner

* Maka `Fee` yang dibebankan ke `Merchant` adalah Rp. 50.000,- * 10% = Rp. 5.000,-
* Dengan demikian, `Platform` menerima total uang sebesar Rp. 50.000,-
* Dimana Rp. 25.000,- akan diberikan ke `User` dan ditanggung oleh pihak `Bank`
* Lalu Rp. 45.000,- akan diberikan ke `Merchant`
* Dan Rp. 5.000,- akan diterima oleh `Platform` sebagai profit margin

### `User` melakukan transaksi di `Merchant` dengan total nilai transaksi = Rp. 20.000,- menggunakan kartu `Bank` partner

* Maka `Fee` yang dibebankan ke `Merchant` adalah Rp. 20.000,- * 10% = Rp. 2.000,-
* Dengan demikian, `Platform` menerima total uang sebesar Rp. 20.000,-
* Dimana Rp. 10.000,- akan diberikan ke `User` dan ditanggung oleh pihak `Bank`
* Dimana Rp. 18.000,- akan diberikan ke `Merchant`
* Dan Rp. 2.000,- akan diterima oleh `Platform` sebagai profit margin

Notice dengan subsidi silang kayak gini, pihak `Platform` selalu untung! Bahkan gak ada impas-impas nya.
Notice juga dimana dengan subsidi silang, pihak `Platform` bisa mencabut aturan minimum transaksi untuk memaksimalkan tanggungan pihak `Bank`.

Crazy right?!?! Tapi kita belum selesai subsidi silangnya...

## Subsidi Merchant

Kalo udah bisa punya skema subsidi silang yang dananya datang dari pihak `Bank`, kenapa gak sekalian aja subsidinya datang dari pihak `Merchant`?
Toh mereka yang jualan. Makin banyak customer juga mereka makin happy, walaopun profit sedikit tergerus.

Maka dateng lagi lah `Promo` dari pihak `Platform` dimana:

* Pihak Platform merumuskan peraturan promo (Terms and Condition bahasa keren-nya)
* Promo akan aktif di event / tanggal tertentu hingga tanggal tertentu, di area tertentu
* Biaya akan ditanggung, sebagian atau **sepenuhnya**, oleh pihak `Merchant`
* Broadcast ke semua `Merchant` di area target, dan minta persetujuan `Merchant`
* Semua `Merchant` yang teken tombol `Setuju` bakal otomatis ikutan `Promo` dan di expose ke banner / list khusus

Lihat gimana pinter-nya pihak platform bikin peraturan yang sangat menguntungkan?
udah 2x kita simulasikan, gw rasa disini udah gak perlu kita simulasi-in lagi karena hasilnya bakal sama aja

> The House Always Wins

## Tapi bro, gak mungkin semua `Platform` pake skema `MDR`

Gw cuma bisa cerita tentang `GoF***` `\(OwO)/` dan `Go***` disini karena gw actually kurang tau case di `Platform` lain gimana.
Jadi mereka punya 2 treatment berbeda untuk transaksi `offline` dan transaksi `online`.

* `Offline` adalah transaksi yang terjadi di lokasi `Merchant`
* `Online` adalah transaksi yang terjadi melalui `GoF***`

Untuk transaksi `offline`, `Go***` akan menanggung semua biaya cashback. Makanya kalo anda notice, biasanya maximal cashback transaksi `offline` ini kecil.

Tapi untuk transaksi `online`, `GoF***` sebagai penyedia platform akan membebankan `MDR` ke `Merchant`.
Nilainya pun gak tanggung-tanggung, sejauh ini gw tau `GoF***` narik `MDR` bisa:

**UP-TO `25%` DARI TOTAL NILAI TRANSAKSI**

## My take-away on all of this

Sebagai `User`, kadang kita gampang merasa diuntungkan oleh penyedia `Platform` yang gencar nyebar promo.
Sebagian besar publik pun senang dengan promo yang diadakan oleh penyedia `Platform`:

> Pake terus bro, selagi mereka bakar duit...

Itu yang sering gw denger dari orang-orang.
Tapi at the end of the day, jangka panjang-nya, gw rasa:

> **ALL THIS SHENANIGAN HAVE TO STOP!**
> **KARENA UJUNG-UJUNGNYA BUKAN PIHAK `PLATFORM` ATAUPUN PIHAK `MERCHANT` YANG MENANGGUNG KERUGIAN, TAPI `USER`!**

`Merchant` pasti menaikkan harga produknya, setinggi atau sedikit lebih tinggi dari `MDR` yang diterapkan pihak `Platform`, untuk meng-compensate profit margin yang tergerus.
Inilah yang gw rasa jadi biang kerok kenaikan harga makanan hingga 30% dalam 2 tahun terakhir (we're in 2020 now).

**Padahal naik gaji aja gak segitu-gitu amat.**
