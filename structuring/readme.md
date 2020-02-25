# Golang, Microservices, dan Monorepo

Setelah gw ngoding pake golang selama 2 tahun terakhir (2018 - 2020), dan menggunakan metode monorepo di `github`.
Gw mendapatkan ide struktur yang menurut gw terbaik, berdasarkan 2 kriteria: `monorepo` dan `microservice`

## Kenapa Microservice

Unit bisnis berkembang/mati dengan cepat, apalagi di fase awal startup.
Dengan menggunakan arsitektur `microservice`, keunggulan-nya adalah:

1. Masing-masing `microservice` dianggap sebagai 1 unit bisnis
2. Gampang meng-spawn `microservice` baru tanpa ada dependency ke service lain
3. Gampang juga meng-take down `microservice`/`unit bisnis` yang gagal tanpa harus refactor/nyenggol code lain

Tapi ada juga kesulitan dengan menggunakan arsitektur `microservice`.

Contoh yang gw temukan adalah `code sharing`, yang mana masalah ini dapat di-tackle dengan `monorepo`

## Kenapa Monorepo

Argumen kenapa mau pake `monorepo` adalah `code sharing`.
Tapi `code sharing` itu sendiri apa sih??? `Gak usah abstract gitu lah....`

Contoh skenario tanpa monorepo:

* Bayangkan suatu e-commerce namanya `TOKOPA'EDI`
* Di e-commerce itu ada 2 `microservice` namanya `payment` dan `invoice`
* Setiap kali `payment` berhasil, harus selalu memberi & mencatat informasi hasil `payment` ke `invoice`

Pada saat `invoice` service menerima informasi `payment`, `invoice` service tidak memiliki `struct` nya `payment`.
Efeknya maka di `invoice` service harus ada duplikasi `struct` `payment`

Dengan monorepo, `struct` `payment` bisa di share code nya accross multiple `microservices`

## Struktur Project

```bash
project
├── cmd                     # adalah directory yang isinya folder-folder microservice
│   ├── '{microservice a}'  # adalah project "microservice a" dan hanya memiliki fungsi "main" untuk startup aplikasi
│   │   └── main.go
│   └── '{microservice b}'  # adalah project "microservice b" juga hanya memiliki fungsi "main" untuk startup aplikasi
│       └── main.go
├── docker                  # adalah directory yang isinya dockerfile dari masing-masing microservice
│   ├── '{microservice a}'
│   └── '{microservice b}'
├── internal                # adalah internal package nya masing-masing microservice yang ada
│   ├── '{microservice a}'    #  code di "internal" package ini tidak boleh di share antar microservice
│   │   ├── rest            # adalah HTTP entry point kalo mau bikin REST API
│   │   ├── grpc            # adalah gRPC entry point kalo mau bikin gRPC
│   │   ├── repo            # adalah repository / konektor ke database
│   │   └── service         # adalah bisnis logic yang di-handle oleh microservice ybs
│   └── '{microservice b}'
│       ├── rest
│       ├── grpc
│       ├── repo
│       └── service
├── pkg                     # adalah directory package yang isinya boleh di share ke seluruh project yang ada di repository
│   └── models              # adalah data model yang biasanya relasi-nya 1:1 ke database
│       ├── something.go
│       └── another.go
└── readme.md
```

Dengan struktur project seperti di atas, maka seluruh object / code yang berada di dalam `pkg`
 bisa di share ke seluruh project yang ada di dalam `monorepo`, tanpa harus ada duplikasi object

CMIIW :v:
