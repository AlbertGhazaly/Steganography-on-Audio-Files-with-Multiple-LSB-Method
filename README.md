# Steganography on Audio Files with Multiple LSB Method

> Hide any file (PDF, TXT, images, documents, …) inside an MP3 and get it back out, using Multiple-LSB steganography with optional Vigenère encryption, through a Go web app.

[Bahasa Indonesia](README.id.md) · ![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go) ![License: MIT](https://img.shields.io/badge/license-MIT-green)

Built for Minor Assignment II (Tucil II) of **IF4020 Kriptografi** (Cryptography), Semester I 2025/2026, Institut Teknologi Bandung (ITB).

## Table of Contents

- [Why this project?](#why-this-project)
- [Features](#features)
- [Quickstart](#quickstart)
- [Usage](#usage)
  - [Web interface](#web-interface)
  - [REST API](#rest-api)
- [Installation](#installation)
- [Tech Stack](#tech-stack)
- [Project Structure](#project-structure)
- [Test Samples](#test-samples)
- [Team](#team)
- [License](#license)

## Why this project?

- **Hide files in MP3 audio.** The secret file is written into the least significant bits (1–4 bits) of the MP3 data.
- **Get the original file back.** Metadata (original filename, file type, size, LSB configuration) is embedded with the payload, so extraction returns the file with its original name and type.
- **More than plain LSB.** You can encrypt the payload with Vigenère, use the key to decide where bits go, check how much fits before embedding, and measure the quality change with PSNR.

## Features

- **Embed** a secret file into an MP3 using LSB with 1–4 bits per byte
- **Extract** the hidden file along with its metadata
- **Vigenère encryption** of the payload before it is embedded
- **Key-based bit positioning** so the embedding is less predictable
- **Capacity calculation**, in bytes and in human-readable form
- **PSNR calculation** comparing the original MP3 with the stego MP3
- **Static web frontend** for uploading files and testing quickly
- **No third-party dependencies.** Only the Go standard library is used.

## Quickstart

Requires Go 1.21 or newer.

```bash
git clone https://github.com/AlbertGhazaly/Steganography-on-Audio-Files-with-Multiple-LSB-Method.git
cd Steganography-on-Audio-Files-with-Multiple-LSB-Method
go build main.go
./main
```

The server starts on port 8080 and prints its endpoints:

```
Steganography Server starting on :8080
API endpoints:
  GET    /api/health
  POST   /api/embed    - Embed secret file into MP3
  POST   /api/extract  - Extract secret file from MP3
  POST   /api/capacity - Calculate MP3 embedding capacity
  POST   /api/psnr     - Calculate PSNR between original and modified MP3
Frontend available at: http://localhost:8080
```

Then open:

- Frontend: http://localhost:8080
- API health check: http://localhost:8080/api/health

The server creates a temporary `./temp` folder to process uploaded files.

## Usage

### Web interface

Open `http://localhost:8080`, upload an MP3 and a secret file, set the LSB bits, the key and the encryption options, then click **Embed** or **Extract**.

### REST API

Base URL: `http://localhost:8080`. All `POST` endpoints take `multipart/form-data`.

| Method | Endpoint        | Description                              |
|--------|-----------------|------------------------------------------|
| GET    | `/api/health`   | Check server status                      |
| POST   | `/api/embed`    | Embed a file into an MP3                 |
| POST   | `/api/extract`  | Extract a file from an MP3               |
| POST   | `/api/capacity` | Calculate embedding capacity             |
| POST   | `/api/psnr`     | Calculate PSNR between original and stego MP3 |

**`POST /api/embed`**

| Field                  | Type                 | Notes                                    |
|------------------------|----------------------|------------------------------------------|
| `mp3_file`             | file                 | Cover MP3                                |
| `secret_file`          | file                 | File to hide                             |
| `key`                  | string               | Required for the `lsb` method            |
| `use_encryption`       | `"true"` / `"false"` | Encrypt the payload with Vigenère        |
| `use_key_for_position` | `"true"` / `"false"` | Use the key to decide bit positions      |
| `method`               | `"lsb"` / `"header"` | Default `lsb`                            |
| `lsb_bits`             | 1–4                  | Default 1                                |

```bash
curl -X POST http://localhost:8080/api/embed \
  -F "mp3_file=@cover.mp3" \
  -F "secret_file=@secret.pdf" \
  -F "key=mysecretkey" \
  -F "use_encryption=true" \
  -F "use_key_for_position=true" \
  -F "lsb_bits=2" \
  -o stego_cover.mp3
```

The response is the stego MP3 (`audio/mpeg`), served as `stego_<original name>.mp3`.

**`POST /api/extract`**

| Field      | Type   | Notes                                                    |
|------------|--------|----------------------------------------------------------|
| `mp3_file` | file   | Stego MP3                                                |
| `key`      | string | Optional; required if encryption was used when embedding |

```bash
curl -X POST http://localhost:8080/api/extract \
  -F "mp3_file=@stego_cover.mp3" \
  -F "key=mysecretkey" \
  -OJ
```

When metadata is available, the response includes these headers: `X-Original-Filename`, `X-File-Type`, `X-Secret-Size`, `X-Used-Encryption`, `X-Used-Key-Position`, `X-LSB-Bits`.

**`POST /api/capacity`**: fields are `mp3_file` (file), `method` (`"lsb"` / `"header"`) and `lsb_bits` (1–4, for `lsb`). Returns JSON with `capacity_bytes`, `capacity_readable` and `method`.

**`POST /api/psnr`**: fields are `original_file` (file) and `modified_file` (file). Returns JSON including `psnr` and `mse`.

## Installation

- **Requirements:** Go 1.21 or newer. There are no third-party dependencies.
- **Build:** `go build main.go` from the project root, then run `./main`.
- **Port:** 8080 (set in `main.go`).

## Tech Stack

- **Backend:** Go 1.21, standard library only (`net/http`, `encoding/json`, …)
- **Frontend:** HTML, CSS and vanilla JavaScript, served statically from `static/`

## Project Structure

```
.
├── main.go
├── go.mod
├── internal/
│   ├── crypto/           # Vigenère encryption
│   ├── handlers/         # HTTP handlers (embed, extract, capacity, psnr, health)
│   ├── middleware/       # CORS
│   ├── models/           # Request/response types
│   └── stego/            # LSB logic, stego header, metadata
├── static/               # Static frontend (HTML, JS)
└── test/                 # Sample test files (MP3 & payloads)
```

## Test Samples

The `test/` directory has a sample MP3 and payloads for each scenario, for example 1- to 4-bit LSB, encryption, key-based positioning, both combined, an oversized payload, and TXT/JPG/DOCX payloads.

## Team

| NIM      | Name                           | GitHub                                           |
|----------|--------------------------------|--------------------------------------------------|
| 13522150 | Albert Ghazaly                 | [@AlbertGhazaly](https://github.com/AlbertGhazaly) |
| 13522158 | Muhammad Rasheed Qais Tandjung | [@trimonuter](https://github.com/trimonuter)     |

## License

Released under the MIT License. See [`LICENSE`](LICENSE).
